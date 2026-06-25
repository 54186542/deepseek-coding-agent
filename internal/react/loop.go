// Package react implements the ReAct (Reasoning + Acting) loop that
// drives the interaction between the user, the LLM, and the file-editing tools.
package react

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/54186542/deepseek-coding-agent/internal/llm"
	"github.com/54186542/deepseek-coding-agent/internal/platform"
	"github.com/54186542/deepseek-coding-agent/internal/tool"
)

const maxIterations = 20

// Loop drives the ReAct cycle.
type Loop struct {
	llmClient   *llm.Client
	toolReg     *tool.Registry
	platform    *platform.Adapter
	messages    []llm.Message
	toolSchemas []llm.ToolSchema
	reader      *bufio.Reader
}

// NewLoop creates a new ReAct loop.
func NewLoop(llmClient *llm.Client, reg *tool.Registry, adapter *platform.Adapter) *Loop {
	return &Loop{
		llmClient: llmClient,
		toolReg:   reg,
		platform:  adapter,
		reader:    bufio.NewReader(os.Stdin),
	}
}

// Run starts the interactive ReAct loop.
func (l *Loop) Run() error {
	fmt.Println("dca — deepseek coding agent")
	fmt.Println("输入你的编程需求，或输入 /exit 退出。")
	fmt.Println()

	// Build tool schemas for the API
	l.buildToolSchemas()

	for {
		// Read user input
		fmt.Print("> ")
		input, err := l.reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}
		input = strings.TrimSpace(input)

		if input == "/exit" || input == "/quit" {
			fmt.Println("再见！")
			break
		}
		if input == "" {
			continue
		}
		if input == "/help" {
			l.printHelp()
			continue
		}

		// Add user message
		l.messages = append(l.messages, llm.Message{
			Role:    "user",
			Content: input,
		})

		// Run the ReAct loop
		if err := l.runCycle(); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loop) runCycle() error {
	for i := 0; i < maxIterations; i++ {
		resp, err := l.llmClient.Chat(l.messages, l.toolSchemas)
		if err != nil {
			return fmt.Errorf("LLM call failed: %w", err)
		}

		if len(resp.Choices) == 0 {
			return fmt.Errorf("no response from LLM")
		}

		choice := resp.Choices[0]
		msg := choice.Message

		// Add assistant message to history
		l.messages = append(l.messages, msg)

		if msg.Content != "" {
			fmt.Println("\n" + msg.Content + "\n")
		}

		// Print token usage
		if resp.Usage.TotalTokens > 0 {
			fmt.Printf("📊 tokens: %d (prompt %d + completion %d)\n",
				resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
		}

		// Check for tool calls
		if len(msg.ToolCalls) == 0 {
			// Model finished responding
			break
		}

		// Process each tool call
		for _, tc := range msg.ToolCalls {
			if err := l.handleToolCall(tc, &i); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Loop) handleToolCall(tc llm.ToolCall, iteration *int) error {
	name := tc.Function.Name
	args := tc.Function.Arguments

	fmt.Printf("🔧 调用工具: %s\n", name)

	result, err := l.toolReg.Execute(name, args)
	if err != nil {
		// Return error to model
		errMsg := err.Error()
		if result != nil && result.Error != "" {
			errMsg = result.Error
		}
		l.messages = append(l.messages, llm.Message{
			Role:       "tool",
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("Error: %s", errMsg),
		})
		return nil
	}

	// If there's a diff and pending write, ask user for confirmation
	if result.Diff != "" {
		fmt.Println("\n📝 修改预览 (diff):")
		fmt.Println(result.Diff)
		fmt.Println()

		if result.PendingPath != "" && result.PendingContent != "" {
			confirmed := l.askConfirmation()
			if confirmed {
				if err := l.platform.WriteFile(result.PendingPath, result.PendingContent); err != nil {
					l.messages = append(l.messages, llm.Message{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    fmt.Sprintf("Error writing file: %s", err.Error()),
					})
					return nil
				}
				fmt.Printf("✅ 已写入: %s\n\n", l.platform.ToSlash(result.PendingPath))

				// Return success to model
				resultData, _ := json.Marshal(result.Data)
				l.messages = append(l.messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf("修改已应用。结果: %s\nDiff:\n%s", string(resultData), result.Diff),
				})
			} else {
				fmt.Println("⏭️ 已跳过，未写入文件\n")
				l.messages = append(l.messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    "用户取消了修改，文件未被更改。",
				})
			}
		}
	} else {
		// Non-edit tool, return result directly
		resultData, _ := json.Marshal(result.Data)
		l.messages = append(l.messages, llm.Message{
			Role:       "tool",
			ToolCallID: tc.ID,
			Content:    string(resultData),
		})
	}

	return nil
}

func (l *Loop) askConfirmation() bool {
	fmt.Print("是否应用以上修改？(Y/n) ")
	input, err := l.reader.ReadString('\n')
	if err != nil {
		return true // default yes on error
	}
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes"
}

func (l *Loop) buildToolSchemas() {
	schemas := l.toolReg.List()
	for _, s := range schemas {
		l.toolSchemas = append(l.toolSchemas, llm.ToolSchema{
			Type: "function",
			Function: llm.Function{
				Name:        s.Name,
				Description: s.Description,
				Parameters:  s.Parameters,
			},
		})
	}
}

func (l *Loop) printHelp() {
	fmt.Println("可用命令:")
	fmt.Println("  /exit, /quit  — 退出")
	fmt.Println("  /help         — 显示此帮助")
	fmt.Println()
	fmt.Println("工具列表:")
	for _, ts := range l.toolSchemas {
		fmt.Printf("  - %s: %s\n", ts.Function.Name, ts.Function.Description)
	}
	fmt.Println()
}

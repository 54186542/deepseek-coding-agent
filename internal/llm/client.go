// Package llm provides the deepseek API client for chat completions
// with function-calling support.
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultBaseURL = "https://api.deepseek.com"

// Client communicates with the deepseek API.
type Client struct {
	APIKey  string
	BaseURL string
	Model   string
	http    *http.Client
}

// Message represents a chat message in the API request/response.
type Message struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolCalls  []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}

// ToolCall represents a function call requested by the model.
type ToolCall struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Function FunctionCall  `json:"function"`
}

// FunctionCall represents the function details in a tool call.
type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolSchema describes a function-calling tool for the API.
type ToolSchema struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function describes a function in the tool schema.
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  any         `json:"parameters"`
}

// ChatRequest is the request body for chat completions.
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []Message     `json:"messages"`
	Tools    []ToolSchema  `json:"tools,omitempty"`
	Stream   bool          `json:"stream"`
}

// ChatResponse is the response from chat completions.
type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice is a single completion choice.
type Choice struct {
	Index        int      `json:"index"`
	Message      Message  `json:"message"`
	FinishReason string   `json:"finish_reason"`
}

// Usage tracks token usage.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewClient creates a new deepseek API client.
func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = "deepseek-chat"
	}
	return &Client{
		APIKey:  apiKey,
		BaseURL: DefaultBaseURL,
		Model:   model,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Chat sends a chat completion request (non-streaming).
func (c *Client) Chat(messages []Message, tools []ToolSchema) (*ChatResponse, error) {
	req := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &chatResp, nil
}

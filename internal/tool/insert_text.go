package tool

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/54186542/deepseek-coding-agent/internal/platform"
	"github.com/54186542/deepseek-coding-agent/internal/safety"
)

// insertTextArgs are the arguments for insert_text.
type insertTextArgs struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Position string `json:"position"` // "before" or "after"
	Content  string `json:"content"`
}

// NewInsertTextTool creates a tool that inserts text before or after a specific line.
//
//	Inserts new content at a specific line position. Line numbers are 1-based.
//	Useful for adding imports, new functions, or new fields.
func NewInsertTextTool(adapter *platform.Adapter, sandbox Sandbox) *Schema {
	return NewSchema(
		"insert_text",
		"Insert new content before or after a specific line number. Line numbers are 1-based. Use for adding imports, functions, or fields at a specific location.",
		NewParam(map[string]ParamProp{
			"path":     StrProp("File path"),
			"line":     IntProp("1-based line number to insert at"),
			"position": EnumProp("Insert before or after the line", []string{"before", "after"}),
			"content":  StrProp("The text to insert"),
		}, []string{"path", "line", "position", "content"}),
		func(raw json.RawMessage) *Result {
			var args insertTextArgs
			if err := ParseArgs(raw, &args); err != nil {
				return ErrResult(err.Error())
			}

			cleanPath, err := sandbox.Clean(args.Path)
			if err != nil {
				return ErrResult(err.Error())
			}

			content, err := adapter.ReadFile(cleanPath)
			if err != nil {
				return ErrResult(err.Error())
			}

			lines := strings.Split(content, "\n")
			totalLines := len(lines)

			if args.Line < 1 || args.Line > totalLines {
				return ErrResult(fmt.Sprintf("line %d out of range (1-%d)", args.Line, totalLines))
			}

			// Ensure content ends with a newline if inserting as a block
			insertLines := strings.Split(args.Content, "\n")

			var newLines []string
			switch args.Position {
			case "before":
				newLines = append(newLines, lines[:args.Line-1]...)
				newLines = append(newLines, insertLines...)
				newLines = append(newLines, lines[args.Line-1:]...)
			case "after":
				newLines = append(newLines, lines[:args.Line]...)
				newLines = append(newLines, insertLines...)
				newLines = append(newLines, lines[args.Line:]...)
			default:
				return ErrResult(fmt.Sprintf("unknown position: %s (use 'before' or 'after')", args.Position))
			}

			newContent := strings.Join(newLines, "\n")
			diff := safety.ComputeDiff(cleanPath, content, newContent)

			return &Result{
				Success: true,
				Data: map[string]any{
					"path":     adapter.ToSlash(cleanPath),
					"line":     args.Line,
					"position": args.Position,
				},
				Diff:           diff.String(),
				PendingPath:    cleanPath,
				PendingContent: newContent,
			}
		},
	)
}

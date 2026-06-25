package tool

import (
	"encoding/json"

	"github.com/54186542/deepseek-coding-agent/internal/platform"
	"github.com/54186542/deepseek-coding-agent/internal/safety"
)

// appendTextArgs are the arguments for append_text.
type appendTextArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Newline bool   `json:"newline,omitempty"`
}

// NewAppendTextTool creates a tool that appends content to the end of a file.
//
//	Adds content at the end of a file. Useful for adding new functions, configurations, or data.
//	If newline is true (default), a newline is added before the appended content if the file doesn't end with one.
func NewAppendTextTool(adapter *platform.Adapter, sandbox Sandbox) *Schema {
	return NewSchema(
		"append_text",
		"Append content to the end of a file. If newline is true (default), ensures the file ends with a newline before appending.",
		NewParam(map[string]ParamProp{
			"path":    StrProp("File path"),
			"content": StrProp("The text to append to the file"),
			"newline": BoolProp("Ensure a trailing newline before appending (default true)"),
		}, []string{"path", "content"}),
		func(raw json.RawMessage) *Result {
			var args appendTextArgs
			if err := ParseArgs(raw, &args); err != nil {
				return ErrResult(err.Error())
			}

			if !args.Newline {
				args.Newline = true
			}

			cleanPath, err := sandbox.Clean(args.Path)
			if err != nil {
				return ErrResult(err.Error())
			}

			content, err := adapter.ReadFile(cleanPath)
			if err != nil {
				return ErrResult(err.Error())
			}

			newContent := content
			if args.Newline && len(content) > 0 && content[len(content)-1] != '\n' {
				newContent += "\n"
			}
			newContent += args.Content

			diff := safety.ComputeDiff(cleanPath, content, newContent)

			return &Result{
				Success: true,
				Data: map[string]any{
					"path": adapter.ToSlash(cleanPath),
				},
				Diff:           diff.String(),
				PendingPath:    cleanPath,
				PendingContent: newContent,
			}
		},
	)
}

package tool

import (
	"encoding/json"
	"strings"

	"github.com/54186542/deepseek-coding-agent/internal/platform"
	"github.com/54186542/deepseek-coding-agent/internal/safety"
)

// editFileArgs are the arguments for edit_file.
type editFileArgs struct {
	Path      string `json:"path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

// NewEditFileTool creates a tool for exact string replacement.
//
//	Replaces the first occurrence of old_string with new_string in the file.
//	This is the safest edit operation because it uses exact text matching.
//	Always read the file first to ensure the old_string exists and is unique.
func NewEditFileTool(adapter *platform.Adapter, sandbox Sandbox) *Schema {
	return NewSchema(
		"edit_file",
		"Replace the FIRST occurrence of old_string with new_string in a file. This is the safest way to make precise edits. Always read the file first to confirm the old_string exists and is unique.",
		NewParam(map[string]ParamProp{
			"path":       StrProp("File path (relative to project root or absolute)"),
			"old_string": StrProp("The exact text to replace (first occurrence only)"),
			"new_string": StrProp("The replacement text"),
		}, []string{"path", "old_string", "new_string"}),
		func(raw json.RawMessage) *Result {
			var args editFileArgs
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

			idx := strings.Index(content, args.OldString)
			if idx < 0 {
				return ErrResult("old_string not found in file. Read the file again to check its current content.")
			}

			newContent := content[:idx] + args.NewString + content[idx+len(args.OldString):]

			diff := safety.ComputeDiff(cleanPath, content, newContent)

			return &Result{
				Success:        true,
				Data: map[string]any{
					"path":        adapter.ToSlash(cleanPath),
					"action":      "edit_file",
					"description": "Replaced first occurrence of old_string with new_string",
				},
				Diff:           diff.String(),
				PendingPath:    cleanPath,
				PendingContent: newContent,
			}
		},
	)
}

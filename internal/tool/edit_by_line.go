package tool

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/54186542/deepseek-coding-agent/internal/platform"
	"github.com/54186542/deepseek-coding-agent/internal/safety"
)

// editByLineArgs are the arguments for edit_by_line.
type editByLineArgs struct {
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line,omitempty"`
	NewContent  string `json:"new_content,omitempty"`
	Action      string `json:"action"` // "replace" or "delete"
}

// NewEditByLineTool creates a tool that edits files by line numbers.
//
//	Replaces or deletes a range of lines. Line numbers are 1-based.
//	Use when you know the exact line numbers to change.
func NewEditByLineTool(adapter *platform.Adapter, sandbox Sandbox) *Schema {
	return NewSchema(
		"edit_by_line",
		"Replace or delete a range of lines by line number. Line numbers are 1-based. For 'replace' action, provide new_content. For 'delete' action, new_content is ignored.",
		NewParam(map[string]ParamProp{
			"path":        StrProp("File path"),
			"start_line":  IntProp("1-based starting line number"),
			"end_line":    IntProp("1-based ending line number (inclusive). If omitted, only start_line is affected."),
			"new_content": StrProp("The new text to replace the line range with (required for 'replace' action)"),
			"action":      EnumProp("Action to perform: 'replace' or 'delete'", []string{"replace", "delete"}),
		}, []string{"path", "start_line", "action"}),
		func(raw json.RawMessage) *Result {
			var args editByLineArgs
			if err := ParseArgs(raw, &args); err != nil {
				return ErrResult(err.Error())
			}

			if args.EndLine == 0 {
				args.EndLine = args.StartLine
			}

			if args.Action == "replace" && args.NewContent == "" {
				return ErrResult("new_content is required for 'replace' action")
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

			if args.StartLine < 1 || args.StartLine > totalLines {
				return ErrResult(fmt.Sprintf("start_line %d out of range (1-%d)", args.StartLine, totalLines))
			}
			if args.EndLine < args.StartLine || args.EndLine > totalLines {
				return ErrResult(fmt.Sprintf("end_line %d out of range (1-%d)", args.EndLine, totalLines))
			}

			var newContent string
			switch args.Action {
			case "delete":
				// Remove lines start..end
				before := lines[:args.StartLine-1]
				after := lines[args.EndLine:]
				newContent = strings.Join(append(before, after...), "\n")
			case "replace":
				before := lines[:args.StartLine-1]
				after := lines[args.EndLine:]
				replacement := args.NewContent
				parts := []string{strings.Join(before, "\n")}
				if replacement != "" {
					parts = append(parts, replacement)
				}
				if len(after) > 0 {
					parts = append(parts, strings.Join(after, "\n"))
				}
				newContent = strings.Join(parts, "\n")
			default:
				return ErrResult(fmt.Sprintf("unknown action: %s (use 'replace' or 'delete')", args.Action))
			}

			diff := safety.ComputeDiff(cleanPath, content, newContent)

			return &Result{
				Success: true,
				Data: map[string]any{
					"path":   adapter.ToSlash(cleanPath),
					"action": args.Action,
					"lines":  fmt.Sprintf("%d-%d", args.StartLine, args.EndLine),
				},
				Diff:           diff.String(),
				PendingPath:    cleanPath,
				PendingContent: newContent,
			}
		},
	)
}

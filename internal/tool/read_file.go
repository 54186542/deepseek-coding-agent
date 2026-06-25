package tool

import (
	"encoding/json"

	"github.com/54186542/deepseek-coding-agent/internal/platform"
)

// readFileArgs are the arguments for read_file.
type readFileArgs struct {
	Path      string `json:"path"`
	StartLine *int   `json:"start_line,omitempty"`
	LineCount *int   `json:"line_count,omitempty"`
}

// NewReadFileTool creates a tool that reads a file's content.
//
//	Reads an entire file or a range of lines. Line numbers are 1-based.
//	Returns the file content as a string. Use this before any edit operation
//	to understand what you're about to change.
func NewReadFileTool(adapter *platform.Adapter, sandbox Sandbox) *Schema {
	return NewSchema(
		"read_file",
		"Read the contents of a file. Returns the full file or a range of lines (1-based). Call this first before making any edits to understand the file content.",
		NewParam(map[string]ParamProp{
			"path":       StrProp("File path (relative to project root or absolute)"),
			"start_line": IntProp("Optional: 1-based starting line number. If omitted, reads from the beginning."),
			"line_count": IntProp("Optional: number of lines to read. If omitted, reads all remaining lines from start_line."),
		}, []string{"path"}),
		func(raw json.RawMessage) *Result {
			var args readFileArgs
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

			lines := splitLines(content)

			start := 1
			if args.StartLine != nil {
				start = *args.StartLine
			}

			if start < 1 || start > len(lines) {
				return ErrResult("start_line out of range")
			}

			end := len(lines)
			if args.LineCount != nil && *args.LineCount > 0 {
				end = start + *args.LineCount - 1
				if end > len(lines) {
					end = len(lines)
				}
			}

			selected := lines[start-1 : end]
			result := stringsJoin(selected, "\n")
			// Append line info
			info := struct {
				Path      string `json:"path"`
				TotalLines int   `json:"total_lines"`
				StartLine int    `json:"start_line"`
				EndLine   int    `json:"end_line"`
				Content   string `json:"content"`
			}{
				Path:       adapter.ToSlash(cleanPath),
				TotalLines: len(lines),
				StartLine:  start,
				EndLine:    end,
				Content:    result,
			}
			return OkResult(info)
		},
	)
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	// Use a simple split
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

func stringsJoin(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	n := len(sep) * (len(elems) - 1)
	for _, e := range elems {
		n += len(e)
	}
	b := make([]byte, n)
	bp := 0
	for i, e := range elems {
		if i > 0 {
			bp += copy(b[bp:], sep)
		}
		bp += copy(b[bp:], e)
	}
	return string(b)
}

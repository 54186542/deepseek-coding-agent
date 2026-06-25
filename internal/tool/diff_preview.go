package tool

import (
	"encoding/json"
	"os"

	"github.com/54186542/deepseek-coding-agent/internal/platform"
)

// diffPreviewArgs are the arguments for diff_preview.
type diffPreviewArgs struct {
	Path string `json:"path,omitempty"`
}

// NewDiffPreviewTool creates a tool that shows a summary of a file.
//
//	Returns file metadata (size, line count) and a preview of the content.
//	Omitting path shows the project root summary.
func NewDiffPreviewTool(adapter *platform.Adapter, sandbox Sandbox) *Schema {
	return NewSchema(
		"diff_preview",
		"Show file metadata and a preview of the file. Use to verify the state of a file before or after edits.",
		NewParam(map[string]ParamProp{
			"path": StrProp("File path. If omitted, shows project root summary."),
		}, []string{}),
		func(raw json.RawMessage) *Result {
			var args diffPreviewArgs
			if err := ParseArgs(raw, &args); err != nil {
				return ErrResult(err.Error())
			}

			targetPath := args.Path
			if targetPath == "" {
				wd, err := adapter.Getwd()
				if err != nil {
					return ErrResult(err.Error())
				}
				targetPath = wd
			}

			cleanPath, err := sandbox.Clean(targetPath)
			if err != nil {
				return ErrResult(err.Error())
			}

			info, err := os.Stat(cleanPath)
			if err != nil {
				return ErrResult("path not accessible: " + err.Error())
			}

			result := map[string]any{
				"path":   adapter.ToSlash(cleanPath),
				"is_dir": info.IsDir(),
				"size":   info.Size(),
			}

			if !info.IsDir() {
				content, err := adapter.ReadFile(cleanPath)
				if err != nil {
					return ErrResult(err.Error())
				}
				lineCount := 0
				for _, c := range content {
					if c == '\n' {
						lineCount++
					}
				}
				result["line_count"] = lineCount

				// Preview: first 20 lines + last 10 lines if file is large
				const maxPreview = 30
				lines := splitLines(content)
				if len(lines) <= maxPreview {
					result["preview"] = content
				} else {
					preview := ""
					for i := 0; i < 20 && i < len(lines); i++ {
						preview += lines[i] + "\n"
					}
					preview += "... (truncated, " + itoa(len(lines)) + " lines total) ...\n"
					for i := len(lines) - 10; i < len(lines); i++ {
						preview += lines[i]
						if i < len(lines)-1 {
							preview += "\n"
						}
					}
					result["preview"] = preview
				}
			}

			return OkResult(result)
		},
	)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

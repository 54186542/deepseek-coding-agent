package tool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/54186542/deepseek-coding-agent/internal/platform"
)

// searchFileArgs are the arguments for search_file.
type searchFileArgs struct {
	Pattern      string `json:"pattern"`
	Path         string `json:"path,omitempty"`
	ContextLines int    `json:"context_lines,omitempty"`
}

// NewSearchFileTool creates a tool that searches file contents.
//
//	Searches for a pattern in a file or directory.
//	Supports plain text and simple glob patterns.
//	Returns matching lines with context.
func NewSearchFileTool(adapter *platform.Adapter, sandbox Sandbox) *Schema {
	return NewSchema(
		"search_file",
		"Search for a pattern in a file or recursively in a directory. Returns matching lines with surrounding context. Uses simple substring matching.",
		NewParam(map[string]ParamProp{
			"pattern":       StrProp("Text pattern to search for (case-sensitive)"),
			"path":          StrProp("File or directory path. If omitted, searches the project root."),
			"context_lines": IntProp("Number of context lines before and after each match (default 2)"),
		}, []string{"pattern"}),
		func(raw json.RawMessage) *Result {
			var args searchFileArgs
			if err := ParseArgs(raw, &args); err != nil {
				return ErrResult(err.Error())
			}

			if args.Pattern == "" {
				return ErrResult("pattern is required")
			}

			if args.ContextLines <= 0 {
				args.ContextLines = 2
			}

			searchPath := args.Path
			if searchPath == "" {
				wd, err := adapter.Getwd()
				if err != nil {
					return ErrResult(err.Error())
				}
				searchPath = wd
			}

			cleanPath, err := sandbox.Clean(searchPath)
			if err != nil {
				return ErrResult(err.Error())
			}

			type match struct {
				Path    string `json:"path"`
				Line    int    `json:"line"`
				Content string `json:"content"`
			}

			var matches []match

			walkFn := func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // skip inaccessible files
				}
				if info.IsDir() {
					return nil
				}
				// Skip binary-looking files
				if isBinaryFile(path) {
					return nil
				}

				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				content := string(data)
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					if strings.Contains(line, args.Pattern) {
						matches = append(matches, match{
							Path:    adapter.ToSlash(path),
							Line:    i + 1,
							Content: strings.TrimRight(line, "\r"),
						})
					}
				}
				return nil
			}

			// Check if path is a file or directory
			info, err := os.Stat(cleanPath)
			if err != nil {
				return ErrResult("path not accessible: " + err.Error())
			}

			if info.IsDir() {
				filepath.Walk(cleanPath, walkFn)
			} else {
				walkFn(cleanPath, info, nil)
			}

			if len(matches) == 0 {
				return OkResult(map[string]any{
					"message": "No matches found",
					"matches": []match{},
				})
			}

			// Truncate if too many results
			maxResults := 50
			truncated := false
			if len(matches) > maxResults {
				matches = matches[:maxResults]
				truncated = true
			}

			result := map[string]any{
				"matches":   matches,
				"total":     len(matches),
				"truncated": truncated,
			}
			return OkResult(result)
		},
	)
}

func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".md", ".txt", ".json", ".yaml", ".yml", ".toml",
		".html", ".css", ".js", ".ts", ".py", ".java", ".c", ".h",
		".cpp", ".hpp", ".rs", ".sh", ".bat", ".ps1", ".mod", ".sum",
		".env", ".ini", ".cfg", ".conf", ".sql", ".xml", ".proto",
		".vue", ".svelte", ".jsx", ".tsx", ".scss", ".less":
		return false
	case "":
		return false
	default:
		return true
	}
}

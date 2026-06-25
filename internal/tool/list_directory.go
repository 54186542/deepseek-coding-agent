package tool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/54186542/deepseek-coding-agent/internal/platform"
)

// listDirArgs are the arguments for list_directory.
type listDirArgs struct {
	Path      string `json:"path,omitempty"`
	Recursive bool   `json:"recursive,omitempty"`
	MaxDepth  int    `json:"max_depth,omitempty"`
}

// NewListDirectoryTool creates a tool that lists a directory.
//
//	Shows files and subdirectories in a given path.
//	Useful for exploring the project structure.
func NewListDirectoryTool(adapter *platform.Adapter, sandbox Sandbox) *Schema {
	return NewSchema(
		"list_directory",
		"List files and directories in a given path. Use to explore the project structure before making changes.",
		NewParam(map[string]ParamProp{
			"path":      StrProp("Directory path (relative to project root or absolute). If omitted, lists the project root."),
			"recursive": BoolProp("If true, list all nested files recursively"),
			"max_depth": IntProp("Maximum directory depth when recursive is true (default 3)"),
		}, []string{}),
		func(raw json.RawMessage) *Result {
			var args listDirArgs
			if err := ParseArgs(raw, &args); err != nil {
				return ErrResult(err.Error())
			}

			dirPath := args.Path
			if dirPath == "" {
				wd, err := adapter.Getwd()
				if err != nil {
					return ErrResult(err.Error())
				}
				dirPath = wd
			}

			cleanPath, err := sandbox.Clean(dirPath)
			if err != nil {
				return ErrResult(err.Error())
			}

			info, err := os.Stat(cleanPath)
			if err != nil {
				return ErrResult("path not accessible: " + err.Error())
			}
			if !info.IsDir() {
				return ErrResult("path is not a directory")
			}

			if args.MaxDepth <= 0 {
				args.MaxDepth = 3
			}

			type entry struct {
				Name  string `json:"name"`
				Path  string `json:"path"`
				IsDir bool   `json:"is_dir"`
				Size  int64  `json:"size,omitempty"`
			}

			var entries []entry

			if args.Recursive {
				baseDepth := strings.Count(cleanPath, string(filepath.Separator))
				filepath.Walk(cleanPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return nil
					}
					depth := strings.Count(path, string(filepath.Separator)) - baseDepth
					if depth > args.MaxDepth {
						if info.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}
					rel, _ := filepath.Rel(cleanPath, path)
					entryPath := adapter.ToSlash(path)
					entries = append(entries, entry{
						Name:  rel,
						Path:  entryPath,
						IsDir: info.IsDir(),
						Size:  func() int64 { if !info.IsDir() { return info.Size() }; return 0 }(),
					})
					return nil
				})
			} else {
				dirEntries, err := os.ReadDir(cleanPath)
				if err != nil {
					return ErrResult(err.Error())
				}
				for _, de := range dirEntries {
					fi, _ := de.Info()
					size := int64(0)
					if fi != nil && !fi.IsDir() {
						size = fi.Size()
					}
					entries = append(entries, entry{
						Name:  de.Name(),
						Path:  adapter.ToSlash(filepath.Join(cleanPath, de.Name())),
						IsDir: de.IsDir(),
						Size:  size,
					})
				}
			}

			// Sort: directories first, then by name
			sort.Slice(entries, func(i, j int) bool {
				if entries[i].IsDir != entries[j].IsDir {
					return entries[i].IsDir
				}
				return entries[i].Name < entries[j].Name
			})

			return OkResult(map[string]any{
				"path":    adapter.ToSlash(cleanPath),
				"entries": entries,
				"count":   len(entries),
			})
		},
	)
}

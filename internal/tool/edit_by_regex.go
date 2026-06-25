package tool

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/54186542/deepseek-coding-agent/internal/platform"
	"github.com/54186542/deepseek-coding-agent/internal/safety"
)

// editByRegexArgs are the arguments for edit_by_regex.
type editByRegexArgs struct {
	Path        string `json:"path"`
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	ReplaceAll  bool   `json:"replace_all,omitempty"`
}

// NewEditByRegexTool creates a tool that edits files using regex.
//
//	Replaces text matching a regular expression. Supports both single and global replacement.
//	Use when you need to change a pattern rather than exact text.
//	The pattern uses Go's regexp syntax (RE2). For literal text replacement, use edit_file instead.
func NewEditByRegexTool(adapter *platform.Adapter, sandbox Sandbox) *Schema {
	return NewSchema(
		"edit_by_regex",
		"Replace text matching a regular expression. Uses Go RE2 syntax. Prefer edit_file for simple literal replacements.",
		NewParam(map[string]ParamProp{
			"path":        StrProp("File path"),
			"pattern":     StrProp("Regular expression pattern (Go RE2 syntax)"),
			"replacement": StrProp("Replacement text (supports $1, $2 etc. for capture groups)"),
			"replace_all": BoolProp("If true, replace ALL occurrences. If false, replace only the first match."),
		}, []string{"path", "pattern", "replacement"}),
		func(raw json.RawMessage) *Result {
			var args editByRegexArgs
			if err := ParseArgs(raw, &args); err != nil {
				return ErrResult(err.Error())
			}

			re, err := regexp.Compile(args.Pattern)
			if err != nil {
				return ErrResult(fmt.Sprintf("invalid regex pattern: %s", err.Error()))
			}

			cleanPath, err := sandbox.Clean(args.Path)
			if err != nil {
				return ErrResult(err.Error())
			}

			content, err := adapter.ReadFile(cleanPath)
			if err != nil {
				return ErrResult(err.Error())
			}

			// Check if pattern matches
			if !re.MatchString(content) {
				return ErrResult("pattern does not match any text in the file. Read the file again to check its content.")
			}

			var newContent string
			if args.ReplaceAll {
				newContent = re.ReplaceAllString(content, args.Replacement)
			} else {
				newContent = re.ReplaceAllString(content, args.Replacement)
				// If replace_all is false but multiple matches exist, only replace first
				// Actually ReplaceAllString replaces all. Let's only replace first.
				loc := re.FindStringIndex(content)
				if loc != nil {
					newContent = content[:loc[0]] + re.ReplaceAllString(content[loc[0]:loc[1]], args.Replacement) + content[loc[1]:]
				}
			}

			diff := safety.ComputeDiff(cleanPath, content, newContent)

			return &Result{
				Success: true,
				Data: map[string]any{
					"path":       adapter.ToSlash(cleanPath),
					"pattern":    args.Pattern,
					"replace_all": args.ReplaceAll,
				},
				Diff:           diff.String(),
				PendingPath:    cleanPath,
				PendingContent: newContent,
			}
		},
	)
}

// Package safety enforces file access boundaries and provides diff-based
// change confirmation so the model cannot accidentally modify files outside
// the project root or in protected directories.
package safety

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Sandbox restricts file operations to a project root directory
// and blocks access to blacklisted paths (e.g. .git, node_modules).
type Sandbox struct {
	ProjectRoot string
	Blacklist   []string
}

// NewSandbox creates a sandbox rooted at projectRoot.
// By default it blocks .git directories.
func NewSandbox(projectRoot string) *Sandbox {
	return &Sandbox{
		ProjectRoot: projectRoot,
		Blacklist:   []string{".git"},
	}
}

// WithBlacklist adds custom blacklisted path components.
func (s *Sandbox) WithBlacklist(patterns []string) *Sandbox {
	s.Blacklist = append(s.Blacklist, patterns...)
	return s
}

// Clean verifies that the given path is within the project root
// and does not match any blacklisted component.
// It returns the cleaned, absolute path on success.
func (s *Sandbox) Clean(path string) (string, error) {
	// Resolve to absolute path
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", path, err)
	}

	// Resolve project root
	rootAbs, err := filepath.Abs(s.ProjectRoot)
	if err != nil {
		return "", fmt.Errorf("resolve project root %s: %w", s.ProjectRoot, err)
	}

	// Must be under project root
	rel, err := filepath.Rel(rootAbs, abs)
	if err != nil {
		return "", fmt.Errorf("rel path %s: %w", path, err)
	}
	if strings.HasPrefix(rel, "..") || rel == "." {
		return "", fmt.Errorf("path %s is outside project root %s", path, s.ProjectRoot)
	}

	// Check blacklist
	components := strings.Split(rel, string(filepath.Separator))
	for _, comp := range components {
		for _, banned := range s.Blacklist {
			if matched, _ := filepath.Match(banned, comp); matched {
				return "", fmt.Errorf("path %s is blocked by blacklist rule %q", path, banned)
			}
		}
	}

	return abs, nil
}

// Diff represents a line-level diff between old and new content.
type Diff struct {
	Path    string
	Hunks   []DiffHunk
	HasDiff bool
}

// DiffHunk is a single changed block.
type DiffHunk struct {
	Action   string // "+" added, "-" removed, " " context
	OldLine  int
	NewLine  int
	Content  string
}

// ComputeDiff computes a simple line diff between old and new content.
// It returns a list of unified-diff-like hunks.
func ComputeDiff(path, oldContent, newContent string) *Diff {
	oldLines := strings.Split(strings.ReplaceAll(oldContent, "\r\n", "\n"), "\n")
	newLines := strings.Split(strings.ReplaceAll(newContent, "\r\n", "\n"), "\n")

	d := &Diff{Path: path}

	// Simple LCS-based diff
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	// Use a simple approach: compare line by line and track differences
	// For simplicity, find the longest common prefix and suffix
	prefixLen := 0
	for prefixLen < len(oldLines) && prefixLen < len(newLines) && oldLines[prefixLen] == newLines[prefixLen] {
		prefixLen++
	}

	suffixLen := 0
	for suffixLen < len(oldLines)-prefixLen && suffixLen < len(newLines)-prefixLen &&
		oldLines[len(oldLines)-1-suffixLen] == newLines[len(newLines)-1-suffixLen] {
		suffixLen++
	}

	oldRemoved := oldLines[prefixLen : len(oldLines)-suffixLen]
	newAdded := newLines[prefixLen : len(newLines)-suffixLen]

	if len(oldRemoved) == 0 && len(newAdded) == 0 {
		d.HasDiff = false
		return d
	}

	d.HasDiff = true

	// Context lines before the change (up to 3)
	ctxBefore := prefixLen - 3
	if ctxBefore < 0 {
		ctxBefore = 0
	}
	for i := ctxBefore; i < prefixLen; i++ {
		d.Hunks = append(d.Hunks, DiffHunk{Action: " ", OldLine: i + 1, NewLine: i + 1, Content: oldLines[i]})
	}

	// Removed lines
	for i, line := range oldRemoved {
		d.Hunks = append(d.Hunks, DiffHunk{Action: "-", OldLine: prefixLen + i + 1, NewLine: 0, Content: line})
	}
	// Added lines
	for i, line := range newAdded {
		d.Hunks = append(d.Hunks, DiffHunk{Action: "+", OldLine: 0, NewLine: prefixLen + i + 1, Content: line})
	}

	// Context lines after the change (up to 3)
	ctxEnd := len(oldLines) - suffixLen + 3
	if ctxEnd > len(oldLines) {
		ctxEnd = len(oldLines)
	}
	suffixStart := len(oldLines) - suffixLen
	if suffixStart < 0 {
		suffixStart = 0
	}
	for i := suffixStart; i < ctxEnd && i < len(oldLines); i++ {
		lineNum := prefixLen + len(oldRemoved) + (i - suffixStart) + 1
		d.Hunks = append(d.Hunks, DiffHunk{Action: " ", OldLine: lineNum, NewLine: lineNum, Content: oldLines[i]})
	}

	return d
}

// String renders the diff in a unified-diff-like format.
func (d *Diff) String() string {
	if !d.HasDiff {
		return "（无变化）"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("--- %s\n+++ %s\n", d.Path, d.Path))
	for _, h := range d.Hunks {
		b.WriteString(fmt.Sprintf("%s %s\n", h.Action, h.Content))
	}
	return b.String()
}

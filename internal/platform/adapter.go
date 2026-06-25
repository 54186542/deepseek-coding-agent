// Package platform provides cross-platform file I/O abstraction.
// It handles path normalization, line ending conversion, and encoding
// so that all tools in this project work identically on Windows, Linux, and macOS.
package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Adapter wraps platform-specific behaviors.
type Adapter struct{}

// NewAdapter creates a new platform adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// NormalizePath converts a raw file path to the platform-native form.
// The tool definitions expose paths as "/" separated for consistency;
// NormalizePath converts them to the OS native separator.
func (a *Adapter) NormalizePath(raw string) string {
	return filepath.FromSlash(raw)
}

// ToSlash converts a platform-native path back to "/" separated form.
// Used when returning paths to the model.
func (a *Adapter) ToSlash(native string) string {
	return filepath.ToSlash(native)
}

// LineEnding returns the native line ending for the current platform.
func (a *Adapter) LineEnding() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

// ReadFile reads a file and normalizes all line endings to "\n".
func (a *Adapter) ReadFile(path string) (string, error) {
	nativePath := a.NormalizePath(path)
	data, err := os.ReadFile(nativePath)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	content := string(data)
	// Normalize \r\n to \n
	content = strings.ReplaceAll(content, "\r\n", "\n")
	return content, nil
}

// WriteFile writes content with the platform-native line ending.
// The content should use "\n" internally; WriteFile converts to native.
func (a *Adapter) WriteFile(path, content string) error {
	nativePath := a.NormalizePath(path)
	// Convert \n to native line ending
	le := a.LineEnding()
	if le != "\n" {
		content = strings.ReplaceAll(content, "\n", le)
	}
	if err := os.WriteFile(nativePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	return nil
}

// FileExists checks whether a file exists and is not a directory.
func (a *Adapter) FileExists(path string) bool {
	nativePath := a.NormalizePath(path)
	info, err := os.Stat(nativePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDir checks whether a path is a directory.
func (a *Adapter) IsDir(path string) bool {
	nativePath := a.NormalizePath(path)
	info, err := os.Stat(nativePath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Getwd returns the current working directory in "/" form.
func (a *Adapter) Getwd() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	return a.ToSlash(wd), nil
}

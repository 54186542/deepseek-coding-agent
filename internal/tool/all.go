package tool

import (
	"github.com/54186542/deepseek-coding-agent/internal/platform"
)

// RegisterAll registers all file operation tools into the given registry.
func RegisterAll(r *Registry, adapter *platform.Adapter, sandbox Sandbox) {
	r.Register(NewReadFileTool(adapter, sandbox))
	r.Register(NewSearchFileTool(adapter, sandbox))
	r.Register(NewListDirectoryTool(adapter, sandbox))
	r.Register(NewEditFileTool(adapter, sandbox))
	r.Register(NewEditByLineTool(adapter, sandbox))
	r.Register(NewEditByRegexTool(adapter, sandbox))
	r.Register(NewInsertTextTool(adapter, sandbox))
	r.Register(NewAppendTextTool(adapter, sandbox))
	r.Register(NewDiffPreviewTool(adapter, sandbox))
}

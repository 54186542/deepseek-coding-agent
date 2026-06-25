// Package tool defines all function-calling tools that the deepseek model
// can invoke to read, search, edit, and manage files in the project.
//
// Each tool is registered as an individual function-calling entry so the
// model can call them independently.
package tool

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Schema describes a function-calling tool for the LLM API.
type Schema struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Parameters  Param   `json:"parameters"`
	Handler     Handler `json:"-"`  // not serialized, used at runtime
}

// Param describes the JSON Schema for a tool's arguments.
type Param struct {
	Type       string                `json:"type"`
	Properties map[string]ParamProp  `json:"properties"`
	Required   []string              `json:"required,omitempty"`
}

// ParamProp describes a single property in the tool's argument schema.
type ParamProp struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
	Required    bool     `json:"-"`
}

// Handler is a function that executes a tool call.
type Handler func(args json.RawMessage) *Result

// Result is the unified return value of a tool execution.
type Result struct {
	Success        bool   `json:"success"`
	Data           any    `json:"data,omitempty"`
	Error          string `json:"error,omitempty"`
	Diff           string `json:"diff,omitempty"`       // present for edit operations
	PendingPath    string `json:"-"`                     // internal: file to write after user confirmation
	PendingContent string `json:"-"`                     // internal: new content to write
}

// Registry holds all registered tools.
type Registry struct {
	tools map[string]*Schema
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]*Schema),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(t *Schema) {
	r.tools[t.Name] = t
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (*Schema, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List returns all tool schemas for function-calling API.
func (r *Registry) List() []Schema {
	list := make([]Schema, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, *t)
	}
	return list
}

// Execute runs a tool by name with the given JSON arguments.
func (r *Registry) Execute(name string, args json.RawMessage) (*Result, error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	result := t.Handler(args)
	if !result.Success {
		return result, fmt.Errorf("%s: %s", name, result.Error)
	}
	return result, nil
}

// Sandbox is the interface for path safety checks.
// Implementations enforce that file operations stay within the project root.
type Sandbox interface {
	Clean(path string) (string, error)
}

// -- helper constructors --

func NewSchema(name, desc string, params Param, handler Handler) *Schema {
	return &Schema{
		Name:        name,
		Description: desc,
		Parameters:  params,
		Handler:     handler,
	}
}

func NewParam(properties map[string]ParamProp, required []string) Param {
	return Param{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

func StrProp(desc string) ParamProp {
	return ParamProp{Type: "string", Description: desc}
}

func IntProp(desc string) ParamProp {
	return ParamProp{Type: "integer", Description: desc}
}

func BoolProp(desc string) ParamProp {
	return ParamProp{Type: "boolean", Description: desc}
}

func EnumProp(desc string, values []string) ParamProp {
	return ParamProp{Type: "string", Description: desc, Enum: values}
}

func ErrResult(msg string) *Result {
	return &Result{Success: false, Error: msg}
}

func OkResult(data any) *Result {
	return &Result{Success: true, Data: data}
}

func OkResultWithDiff(data any, diff string) *Result {
	return &Result{Success: true, Data: data, Diff: diff}
}

// -- arg helpers --

// ParseArgs unmarshals JSON arguments into the target struct.
// deepseek API returns arguments as a JSON string (e.g. "{\"path\":\".\"}"),
// so we handle both raw objects and string-encoded JSON.
func ParseArgs(raw json.RawMessage, target any) error {
	// If the raw message is a JSON string, unmarshal it first
	var s string
	if json.Unmarshal(raw, &s) == nil {
		// It's a string, use the decoded value
		raw = json.RawMessage(s)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("invalid arguments: %w (raw: %s)", err, string(raw))
	}
	return nil
}

// RequiredFields checks that required string fields are non-empty.
func RequiredFields(args map[string]any, fields ...string) error {
	var missing []string
	for _, f := range fields {
		if v, ok := args[f]; !ok || v == nil || fmt.Sprint(v) == "" {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

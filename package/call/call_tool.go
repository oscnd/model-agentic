package call

import "fmt"

type Tool struct {
	Type        *string `json:"type,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	InputSchema *Schema `json:"inputSchema,omitempty"`
}

type ToolCall struct {
	Id        *string `json:"id"`
	Type      *string `json:"type"`
	Name      *string `json:"name,omitempty"`
	Arguments []byte  `json:"arguments,omitempty"`
	Result    []byte  `json:"output,omitempty"`
}

func (r *ToolCall) String() string {
	return fmt.Sprintf("Name: %s, Request: %s, Response: %s", *r.Name, r.Arguments, r.Result)
}

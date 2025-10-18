package call

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
	Arguments any     `json:"arguments,omitempty"`
	Output    any     `json:"output,omitempty"`
}

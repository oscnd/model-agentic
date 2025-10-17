package call

import "encoding/json"

type Tool struct {
	Type        string          `json:"type,omitempty"`
	Function    *ToolFunction   `json:"function,omitempty"`
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
}

type ToolCall struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Function *ToolCallFunction `json:"function"`
}

type ToolFunction struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
}

type ToolChoice struct {
	Type     string              `json:"type"`
	Name     *string             `json:"name,omitempty"`
	Function *ToolChoiceFunction `json:"function,omitempty"`
}

type ToolChoiceFunction struct {
	Name string `json:"name"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

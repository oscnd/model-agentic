package agent

import "go.scnd.dev/open/model/agentic/package/call"

type Option struct {
	Name                   *string      `json:"name" validate:"required,slug"`
	Persona                *string      `json:"persona" validate:"required"`
	Description            *string      `json:"description" validate:"required"`
	Model                  *string      `json:"model" validate:"required"`
	MaxTokens              *int         `json:"maxTokens"`
	Temperature            *float64     `json:"temperature"`
	TopP                   *float64     `json:"topP"`
	TopK                   *int         `json:"topK"`
	AllowSubagentDispatch  *bool        `json:"allowSubagentDispatch"`
	SubagentDispatchPrompt *string      `json:"subagentDispatchPrompt"`
	SubagentDispatchLimit  *int         `json:"subagentDispatchLimit"`
	CallOption             *call.Option `json:"-"`
}

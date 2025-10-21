package agent

import (
	"go.scnd.dev/open/model/agentic/package/function"
)

// Option contains configuration for agent, extended function options.
type Option struct {
	Name                   *string          `json:"name" validate:"required,slug"`
	Persona                *string          `json:"persona" validate:"required"`
	Description            *string          `json:"description" validate:"required"`
	AllowSubagentDispatch  *bool            `json:"allowSubagentDispatch"`
	SubagentDispatchPrompt *string          `json:"subagentDispatchPrompt"`
	SubagentDispatchLimit  *int             `json:"subagentDispatchLimit"`
	FunctionOption         *function.Option `json:"functionOption"`
}

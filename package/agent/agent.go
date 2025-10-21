// Package agent provides agent implementation with function calling capabilities,
// it handles subagent dispatch, context management, and state-based execution.
package agent

import (
	"go.scnd.dev/open/model/agentic/package/call"
	"go.scnd.dev/open/model/agentic/package/function"
)

// Agent represents an agent with functions and subagents
type Agent struct {
	Caller    call.Caller             `json:"-"`
	Option    *Option                 `json:"option"`
	Functions []*function.Declaration `json:"functions"`
	Subagents []*Agent                `json:"subagents"`
	Messages  []*call.Message         `json:"messages"`
}

func New(caller call.Caller, option *Option) *Agent {
	return &Agent{
		Caller:    caller,
		Option:    option,
		Functions: make([]*function.Declaration, 0),
		Subagents: make([]*Agent, 0),
		Messages:  make([]*call.Message, 0),
	}
}

// AddFunction adds a function declaration to the agent
func (r *Agent) AddFunction(declaration *function.Declaration) {
	r.Functions = append(r.Functions, declaration)
}

// AddSubagent adds a subagent to the agent
func (r *Agent) AddSubagent(subagent *Agent) {
	r.Subagents = append(r.Subagents, subagent)
}

package agent

import (
	"go.scnd.dev/open/model-agentic/package/call"
	"go.scnd.dev/open/model-agentic/package/function"
)

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
		Subagents: make([]*Agent, 0),
	}
}

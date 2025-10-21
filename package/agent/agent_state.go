package agent

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
	"go.scnd.dev/open/model/agentic/package/function"
)

type State struct {
	Task          *string         `json:"task"`
	FunctionState *function.State `json:"-"`
}

func (r *Agent) NewState(task *string) *State {
	// * construct messages
	messages := make([]*call.Message, 0)

	// * create system message
	systemMessage := &call.Message{
		Role:      gut.Ptr("system"),
		Content:   r.Option.Persona,
		Images:    nil,
		ToolCalls: nil,
	}

	// * create user message
	userMessage := &call.Message{
		Role:      gut.Ptr("user"),
		Content:   task,
		Images:    nil,
		ToolCalls: nil,
	}

	messages = append(messages, systemMessage, userMessage)
	messages = append(messages, r.Messages...)

	// * create and return state
	return &State{
		Task:          task,
		FunctionState: function.NewState(messages),
	}
}

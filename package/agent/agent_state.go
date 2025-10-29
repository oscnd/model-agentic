package agent

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
	"go.scnd.dev/open/model/agentic/package/function"
)

// State manages the execution state of an agent with task and function state
type State struct {
	Task          *string         `json:"task"`
	FunctionState *function.State `json:"-"`
}

// NewState creates a new agent state with the specified task and initial messages
func (r *Agent) NewState(task *string) *State {
	// * construct messages
	messages := make([]*call.Message, 0)

	// * create system message
	systemMessage := &call.Message{
		Role:        gut.Ptr("system"),
		Content:     r.Option.Persona,
		Image:       nil,
		ImageUrl:    nil,
		ImageDetail: nil,
		ToolCalls:   nil,
		Usage:       nil,
	}

	// * create user message
	userMessage := &call.Message{
		Role:        gut.Ptr("user"),
		Content:     task,
		Image:       nil,
		ImageUrl:    nil,
		ImageDetail: nil,
		ToolCalls:   nil,
		Usage:       nil,
	}

	messages = append(messages, systemMessage, userMessage)
	messages = append(messages, r.Messages...)

	// * create and return state
	return &State{
		Task:          task,
		FunctionState: function.NewState(messages),
	}
}

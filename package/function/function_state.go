package function

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
)

type StateOnBeforeFunctionCall func(callback *CallbackBeforeFunctionCall) (map[string]any, *gut.ErrorInstance)

type StateOnAfterFunctionCall func(callback *CallbackAfterFunctionCall) (map[string]any, *gut.ErrorInstance)

type StateOnToolMessage func(message *call.Message) *gut.ErrorInstance

// State uses for manages conversation messages and callback hooks using function calling
type State struct {
	InitialMessages      []*call.Message           `json:"initialMessages"`
	ToolMessages         []*call.Message           `json:"toolMessages"`
	OnBeforeFunctionCall StateOnBeforeFunctionCall `json:"-"`
	OnAfterFunctionCall  StateOnAfterFunctionCall  `json:"-"`
	OnToolMessage        StateOnToolMessage        `json:"-"`
}

// NewState creates a new function calling state with initial messages
func NewState(initialMessages []*call.Message) *State {
	return &State{
		InitialMessages: initialMessages,
		ToolMessages:    make([]*call.Message, 0),
	}
}

// Messages returns all messages in chronological order
func (r *State) Messages() []*call.Message {
	messages := make([]*call.Message, len(r.InitialMessages)+len(r.ToolMessages))
	messages = append(messages, r.InitialMessages...)
	messages = append(messages, r.ToolMessages...)
	return messages
}

// Inherit copies callback hooks from another state
func (r *State) Inherit(state *State) {
	r.OnBeforeFunctionCall = state.OnBeforeFunctionCall
	r.OnAfterFunctionCall = state.OnAfterFunctionCall
	r.OnToolMessage = state.OnToolMessage
}

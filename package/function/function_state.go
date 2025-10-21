package function

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
)

type StateOnBeforeFunctionCall func(callback *CallbackBeforeFunctionCall) (map[string]any, *gut.ErrorInstance)
type StateOnAfterFunctionCall func(callback *CallbackAfterFunctionCall) (map[string]any, *gut.ErrorInstance)
type StateOnToolMessage func(message *call.Message) *gut.ErrorInstance

type State struct {
	InitialMessages      []*call.Message           `json:"initialMessages"`
	ToolMessages         []*call.Message           `json:"toolMessages"`
	OnBeforeFunctionCall StateOnBeforeFunctionCall `json:"-"`
	OnAfterFunctionCall  StateOnAfterFunctionCall  `json:"-"`
	OnToolMessage        StateOnToolMessage        `json:"-"`
}

func NewState(initialMessages []*call.Message) *State {
	return &State{
		InitialMessages: initialMessages,
		ToolMessages:    make([]*call.Message, 0),
	}
}

func (r *State) Messages() []*call.Message {
	messages := make([]*call.Message, len(r.InitialMessages)+len(r.ToolMessages))
	messages = append(messages, r.InitialMessages...)
	messages = append(messages, r.ToolMessages...)
	return messages
}

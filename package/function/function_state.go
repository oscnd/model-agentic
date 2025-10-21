package function

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
)

type StateBeforeFunctionCall func(callback *CallbackBeforeFunctionCall) (map[string]any, *gut.ErrorInstance)
type StateAfterFunctionCall func(callback *CallbackAfterFunctionCall) (map[string]any, *gut.ErrorInstance)

type State struct {
	InitialMessages      []*call.Message         `json:"initialMessages"`
	ToolMessages         []*call.Message         `json:"toolMessages"`
	OnBeforeFunctionCall StateBeforeFunctionCall `json:"-"`
	OnAfterFunctionCall  StateAfterFunctionCall  `json:"-"`
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

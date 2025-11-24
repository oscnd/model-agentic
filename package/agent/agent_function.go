package agent

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
	"go.scnd.dev/open/model/agentic/package/function"
)

// Function creates a function declaration for the agent to be used for executing from a parent agent
func (r *Agent) Function(state *State) *function.Declaration {
	type Arguments struct {
		Task           *string `json:"task" description:"The task or question to be processed by the subagent" validate:"required"`
		IncludeContext *bool   `json:"includeContext" description:"Whether to include the parent agent's context to subagent" validate:"required"`
	}

	declaration := function.NewDeclaration(
		gut.Ptr("call_"+*r.Option.Name),
		r.Option.Description,
		func(arguments *Arguments) (map[string]any, *gut.ErrorInstance) {
			// * validate arguments
			if arguments.Task == nil {
				return nil, gut.Err(false, "task arguments is required", nil)
			}

			agent := New(r.Caller, r.Option)
			agent.Functions = r.Functions
			agent.Subagents = r.Subagents
			agentState := agent.NewState(arguments.Task)
			agentState.FunctionState.Inherit(state.FunctionState)

			// * include context from parent state
			if arguments.IncludeContext != nil && *arguments.IncludeContext && state != nil && state.FunctionState != nil {
				messages := state.FunctionState.Messages()
				for _, message := range messages {
					switch message.(type) {
					case *call.SystemMessage:
						m := message.(*call.SystemMessage)
						agent.ContextPush(*m.Content)
					case *call.UserMessage:
						m := message.(*call.UserMessage)
						agent.ContextPush(*m.Content)
					case *call.AssistantMessage:
						m := message.(*call.AssistantMessage)
						if m.Content != nil {
							agent.ContextPush(*m.Content)
						}
						for _, toolCall := range m.ToolCalls {
							agent.ContextPush(toolCall.String())
						}
					}
				}
			}

			response, err := agent.Call(agentState, nil)
			if err != nil {
				return nil, gut.Err(false, "agent function call error: "+err.Error(), err)
			}
			return map[string]any{
				"response": response.Message.Content,
			}, nil
		},
	)

	if r.Option.Terminator != nil && *r.Option.Terminator {
		declaration.Terminator = r.Option.Terminator
	}

	return declaration
}

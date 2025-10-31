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

	return &function.Declaration{
		Name:        gut.Ptr("call_" + *r.Option.Name),
		Description: r.Option.Description,
		Argument:    call.SchemaConvert(new(Arguments)),
		Func: func(args map[string]any) (map[string]any, *gut.ErrorInstance) {
			// * validate arguments
			taskRaw, exists := args["task"]
			if !exists {
				return nil, gut.Err(false, "task argument is required", nil)
			}
			task, ok := taskRaw.(string)
			if !ok {
				return nil, gut.Err(false, "task must be a string", nil)
			}

			includeContextRaw, exists := args["includeContext"]
			if !exists {
				return nil, gut.Err(false, "includeContext argument is required", nil)
			}
			includeContext, ok := includeContextRaw.(bool)
			if !ok {
				return nil, gut.Err(false, "includeContext must be a boolean", nil)
			}

			agent := New(r.Caller, r.Option)
			agent.Functions = r.Functions
			agent.Subagents = r.Subagents
			agentState := agent.NewState(&task)
			agentState.FunctionState.Inherit(state.FunctionState)

			// * include context from parent state
			if includeContext && state != nil && state.FunctionState != nil {
				messages := state.FunctionState.Messages()
				for _, m := range messages {
					if m.Content != nil && *m.Role != "system" {
						agent.ContextPush(*m.Content)
					}
					for _, toolCall := range m.ToolCalls {
						agent.ContextPush(toolCall.String())
					}
				}
			}

			var output map[string]any
			if _, err := agent.Call(agentState, &output); err != nil {
				return nil, gut.Err(false, "agent function call error: "+err.Error(), err)
			}
			return output, nil
		},
	}
}

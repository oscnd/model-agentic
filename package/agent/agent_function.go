package agent

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model-agentic/package/call"
	"go.scnd.dev/open/model-agentic/package/function"
)

func (r *Agent) Function(caller function.Caller, callback func(invoke *function.CallbackInvoke)) *function.Declaration {
	type Arguments struct {
		Task           *string `json:"task" description:"The task or question to be processed by the subagent" validate:"required"`
		IncludeContext *bool   `json:"includeContext" description:"Whether to include the parent agent's context to subagent" validate:"required"`
	}

	type Result struct {
		Output *string `json:"output" description:"The output from the task processed"`
	}

	return &function.Declaration{
		Name:        gut.Ptr("call_" + *r.Option.Name),
		Description: r.Option.Description,
		Argument:    call.SchemaConvert(new(Arguments)),
		Func: func(args map[string]any) (map[string]any, *gut.ErrorInstance) {
			task := args["task"].(string)
			includeContext := args["includeContext"].(bool)
			agent := New(r.Caller, r.Option)

			if includeContext {
				for _, m := range caller.Message() {
					if m.Content != nil && *m.Role != "system" {
						agent.ContextPush(*m.Content)
					}
					for _, toolCall := range m.ToolCalls {
						agent.ContextPush(toolCall.String())
					}
				}
			}

			output := new(Result)
			if _, err := agent.Call(&task, &output, callback); err != nil {
				return nil, gut.Err(false, "agent function call error", err)
			}
			return map[string]any{
				"output": output.Output,
			}, nil
		},
	}
}

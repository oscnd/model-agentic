package agent

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model-agentic/package/call"
	"go.scnd.dev/open/model-agentic/package/function"
)

func (r *Agent) Function(callback func(invoke *function.CallbackInvoke)) *function.Declaration {
	type Arguments struct {
		Task *string `json:"task"`
	}

	return &function.Declaration{
		Name:        gut.Ptr("call_" + *r.Option.Name),
		Description: r.Option.Description,
		Argument:    call.SchemaConvert(new(Arguments)),
		Func: func(args map[string]any) (map[string]any, *gut.ErrorInstance) {
			task := args["task"].(string)
			var output map[string]any
			if _, err := r.Call(&task, &output, callback); err != nil {
				return nil, gut.Err(false, "agent function call error", err)
			}
			return output, nil
		},
	}
}

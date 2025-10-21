package agent

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
	"go.scnd.dev/open/model/agentic/package/function"
)

func (r *Agent) Call(state *State, output any) (*call.Response, *gut.ErrorInstance) {
	// * construct function caller
	caller := function.New(r.Caller, r.Option.FunctionOption)

	// * add function declarations
	functions := append(r.Functions)

	// * add subagent functions
	for _, subagent := range r.Subagents {
		functions = append(functions, subagent.Function(state))
	}

	// * apply functions
	for _, f := range functions {
		caller.AddDeclaration(f)
	}

	// TODO: Add dispatch subagent function

	// * call function caller
	response, err := caller.Call(state.FunctionState, output)
	if err != nil {
		return nil, err
	}

	return response, nil
}

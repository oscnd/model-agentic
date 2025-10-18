package agent

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model-agentic/package/call"
	"go.scnd.dev/open/model-agentic/package/function"
)

func (r *Agent) Call(task *string, output any, callback func(invoke *function.CallbackInvoke)) (*call.Response, *gut.ErrorInstance) {
	// * construct function request
	request := r.CallCreateRequest(task)

	// * add function declarations
	functions := append(r.Functions)

	// * add subagent functions
	for _, subagent := range r.Subagents {
		functions = append(functions, subagent.Function(callback))
	}

	// * apply functions
	caller := function.New(r.Caller)
	for _, f := range functions {
		caller.AddDeclaration(f)
	}

	// TODO: Add dispatch subagent function

	// * call function caller
	response, err := caller.Call(request, r.Option.CallOption, output, callback)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (r *Agent) CallCreateRequest(task *string) *function.Request {
	// * create function request
	request := &function.Request{
		Model:       r.Option.Model,
		Messages:    nil,
		MaxTokens:   r.Option.MaxTokens,
		Temperature: r.Option.Temperature,
		TopP:        r.Option.TopP,
		TopK:        r.Option.TopK,
	}

	// * create message
	systemMessage := &call.Message{
		Role:      gut.Ptr("system"),
		Content:   r.Option.Persona,
		Images:    nil,
		ToolCalls: nil,
	}

	userMessage := &call.Message{
		Role:      gut.Ptr("user"),
		Content:   task,
		Images:    nil,
		ToolCalls: nil,
	}

	request.Messages = append(request.Messages, systemMessage, userMessage)

	return request
}

package main

import (
	"os"

	"github.com/bsthun/gut"
	"github.com/davecgh/go-spew/spew"
	"go.scnd.dev/open/model/agentic/package/call"
	"go.scnd.dev/open/model/agentic/package/function"
)

func main() {
	// * initialize caller
	caller := call.NewOpenai(os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))
	model := os.Getenv("OPENAI_MODEL")
	maxTokens := 2000
	temperature := 0.7

	// * create function call option
	option := &function.Option{
		Model:       &model,
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
		CallOption:  new(call.Option),
	}

	// * create function call instance
	functionCall := function.New(caller, option)

	// * fetch MCP server declarations from Model Context Protocol
	mcpOption := &function.McpOption{
		BaseUrl: "https://modelcontextprotocol.io/mcp",
		Header: map[string]string{
			"Authorization": "Bearer X",
		},
	}

	// * get mcp declarations
	declarations, err := function.McpDeclarations(mcpOption)
	if err != nil {
		gut.Fatal("failed to fetch mcp declarations", err)
	} else {
		// * add mcp declarations to function call
		for _, declaration := range declarations {
			functionCall.AddDeclaration(declaration)
		}
	}

	// * create state with initial messages
	state := function.NewState([]call.Message{
		&call.UserMessage{
			Content: gut.Ptr("What is the architecture of Model Context Protocol (MCP) and what are some examples of remote server implementations? Please use the available tools to gather information and provide a comprehensive answer."),
		},
	})

	// * track invocation using callbacks
	var beforeInvocations []*function.CallbackBeforeFunctionCall
	var afterInvocations []*function.CallbackAfterFunctionCall

	state.OnBeforeFunctionCall = func(callback *function.CallbackBeforeFunctionCall) (map[string]any, *gut.ErrorInstance) {
		beforeInvocations = append(beforeInvocations, callback)
		return nil, nil
	}

	state.OnAfterFunctionCall = func(callback *function.CallbackAfterFunctionCall) (map[string]any, *gut.ErrorInstance) {
		afterInvocations = append(afterInvocations, callback)
		return nil, nil
	}

	// * call function
	response, er := functionCall.Call(state, nil)
	if er != nil {
		gut.Fatal("function call failed", err)
	}

	// * display results
	println("=== MCP Function Call Example Results ===")
	println("\n1. Final Response:")
	spew.Dump(response)

	println("\n2. MCP Tool Invocation:")
	for _, invoke := range afterInvocations {
		spew.Dump(invoke)
	}

	println("\n4. Total Function Invocations:", len(afterInvocations))
}

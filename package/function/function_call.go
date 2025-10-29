// Package function provides function calling capabilities with state management,
// it handles function declarations, looped function calling execution, and callback hooks.
package function

import (
	"encoding/json"

	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
)

// Caller defines the interface for function calling
type Caller interface {
	// AddDeclaration registers a new function declaration
	AddDeclaration(declaration *Declaration)
	// Call executes function calls with state management and callbacks
	Call(state *State, output any) (*call.Response, *gut.ErrorInstance)
}

type Call struct {
	Caller       call.Caller    `json:"-"`
	Option       *Option        `json:"option"`
	Declarations []*Declaration `json:"declarations"`
}

func New(caller call.Caller, option *Option) Caller {
	return &Call{
		Caller:       caller,
		Option:       option,
		Declarations: make([]*Declaration, 0),
	}
}

// AddDeclaration appends a new function declaration to the registry
func (r *Call) AddDeclaration(declaration *Declaration) {
	r.Declarations = append(r.Declarations, declaration)
}

// Call executes the function calling loop with state management and callbacks
// during execution, state passed can use to get current messages and set callbacks
func (r *Call) Call(state *State, output any) (*call.Response, *gut.ErrorInstance) {
	// * convert function request to call request by appending function declarations as tools
	callRequest := &call.Request{
		Model:       r.Option.Model,
		MaxTokens:   r.Option.MaxTokens,
		Temperature: r.Option.Temperature,
		TopP:        r.Option.TopP,
		TopK:        r.Option.TopK,
		Messages:    nil,
		Tools:       r.Tools(),
	}

	// * loop until no more tool calls
	for {
		// * call underlying caller
		callRequest.Messages = state.Messages()
		response, err := r.Caller.Call(callRequest, r.Option.CallOption, output)
		if err != nil {
			return nil, err
		}

		// * check if there are tool calls
		if response.FinishReason != "tool_calls" && len(response.Message.ToolCalls) == 0 {
			// * append final message
			callRequest.Messages = append(callRequest.Messages, response.Message)

			// * aggregate usage from all messages
			for _, message := range callRequest.Messages {
				if message.Usage == nil {
					continue
				}
				*response.Usage.InputTokens += *message.Usage.InputTokens
				*response.Usage.OutputTokens += *message.Usage.OutputTokens
				*response.Usage.CachedTokens += *message.Usage.CachedTokens
			}

			return response, nil
		}

		// * process each tool call
		toolCalls := make([]*call.ToolCall, 0)
		for _, toolCall := range response.Message.ToolCalls {
			// * find matching declaration
			declaration := r.GetDeclaration(toolCall.Name)
			if declaration == nil {
				return nil, gut.Err(false, "declaration not found for tool: "+gut.Val(toolCall.Name), nil)
			}

			// * unmarshal arguments from json
			var arguments map[string]any
			if err := json.Unmarshal(toolCall.Arguments, &arguments); err != nil {
				return nil, gut.Err(false, "failed to unmarshal tool call arguments", err)
			}

			// * invoke callback before execution with response as nil
			callback := &CallbackBeforeFunctionCall{
				ToolCallId:  toolCall.Id,
				Declaration: declaration,
				Argument:    arguments,
			}
			if state.OnBeforeFunctionCall != nil {
				alter, err := state.OnBeforeFunctionCall(callback)
				if alter != nil {
					arguments = alter
				}
				if err != nil {
					return nil, err
				}
			}

			// * execute function to get response
			functionResponse, funcErr := declaration.Func(arguments)
			if funcErr != nil {
				return nil, funcErr
			}

			// * invoke callback after execution with response
			if state.OnAfterFunctionCall != nil {
				alter, err := state.OnAfterFunctionCall(&CallbackAfterFunctionCall{
					CallbackBeforeFunctionCall: *callback,
					Result:                     functionResponse,
				})
				if alter != nil {
					functionResponse = alter
				}
				if err != nil {
					return nil, err
				}
			}

			// * marshal response to json
			responseJson, err := json.Marshal(functionResponse)
			if err != nil {
				return nil, gut.Err(false, "failed to marshal function response to json", err)
			}

			// * create tool result message
			toolCall.Result = responseJson
			toolCalls = append(toolCalls, toolCall)
		}

		toolMessage := &call.Message{
			Role:        gut.Ptr("tool"),
			Content:     response.Message.Content,
			Image:       nil,
			ImageUrl:    nil,
			ImageDetail: nil,
			ToolCalls:   toolCalls,
			Usage:       response.Usage,
		}

		// * append result message
		state.ToolMessages = append(state.ToolMessages, toolMessage)
		if state.OnToolMessage != nil {
			if err := state.OnToolMessage(toolMessage); err != nil {
				return nil, err
			}
		}
	}
}

// Tools converts function declarations to call.Tool format
func (r *Call) Tools() []*call.Tool {
	var tools []*call.Tool
	for _, declaration := range r.Declarations {
		tool := &call.Tool{
			Type:        gut.Ptr("function"),
			Name:        declaration.Name,
			Description: declaration.Description,
			InputSchema: declaration.Argument,
		}
		tools = append(tools, tool)
	}
	return tools
}

// GetDeclaration finds a function declaration by name
func (r *Call) GetDeclaration(name *string) *Declaration {
	if name == nil {
		return nil
	}
	for _, declaration := range r.Declarations {
		if declaration.Name != nil && *declaration.Name == *name {
			return declaration
		}
	}
	return nil
}

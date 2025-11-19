// Package function provides function calling capabilities with state management,
// it handles function declarations, looped function calling execution, and callback hooks.
package function

import (
	"encoding/json"
	"fmt"
	"reflect"

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
		Model:           r.Option.Model,
		MaxTokens:       r.Option.MaxTokens,
		Temperature:     r.Option.Temperature,
		TopP:            r.Option.TopP,
		TopK:            r.Option.TopK,
		ReasoningEffort: r.Option.ReasoningEffort,
		Messages:        nil,
		Tools:           r.Tools(),
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

			// * construct usage field
			response.TotalUsage = &call.Usage{
				InputTokens:  gut.Ptr[int64](0),
				OutputTokens: gut.Ptr[int64](0),
				CachedTokens: gut.Ptr[int64](0),
			}

			// * aggregate usage from all messages
			for _, message := range callRequest.Messages {
				m, ok := message.(*call.AssistantMessage)
				if !ok {
					continue
				}
				if m.Usage == nil {
					continue
				}
				*response.TotalUsage.InputTokens += gut.Val(m.Usage.InputTokens)
				*response.TotalUsage.OutputTokens += gut.Val(m.Usage.OutputTokens)
				*response.TotalUsage.CachedTokens += gut.Val(m.Usage.CachedTokens)
			}

			return response, nil
		}

		// * process each tool call
		toolCalls := make([]*call.ToolCall, 0)
		for _, toolCall := range response.Message.ToolCalls {
			// * find matching declaration
			declaration := r.GetDeclaration(toolCall.Name)
			if declaration == nil {
				if r.Option.ParseErrorBreak != nil && *r.Option.ParseErrorBreak {
					return nil, gut.Err(false, "declaration not found for tool: "+gut.Val(toolCall.Name), nil)
				}
				toolCall.Error = gut.Ptr("declaration not found for tool: " + gut.Val(toolCall.Name))
				toolCalls = append(toolCalls, toolCall)
				continue
			}

			// * unmarshal arguments from json
			elem := reflect.TypeOf(declaration.Arguments).Elem()
			for elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			arguments := reflect.New(elem).Interface()
			if elem.Kind() != reflect.Interface {
				if err := json.Unmarshal(toolCall.Arguments, arguments); err != nil {
					if r.Option.ParseErrorBreak != nil && *r.Option.ParseErrorBreak {
						return nil, gut.Err(false, fmt.Sprintf("failed to unmarshal arguments for tool %s: %s", gut.Val(toolCall.Name), err.Error()), err)
					}
					toolCall.Error = gut.Ptr("failed to unmarshal arguments: " + err.Error())
					toolCalls = append(toolCalls, toolCall)
					continue
				}
			}

			// * invoke callback before execution with response as nil
			callback := &CallbackBeforeFunctionCall{
				ToolCallId:  toolCall.Id,
				Declaration: declaration,
				Arguments:   arguments,
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
				if r.Option.ParseErrorBreak != nil && *r.Option.ParseErrorBreak {
					return nil, gut.Err(false, "function execution error for tool "+gut.Val(toolCall.Name)+": "+funcErr.Error(), funcErr)
				}
				toolCall.Error = gut.Ptr("function execution error: " + funcErr.Error())
				toolCalls = append(toolCalls, toolCall)

				if state.OnAfterFunctionCall != nil {
					alter, err := state.OnAfterFunctionCall(&CallbackAfterFunctionCall{
						CallbackBeforeFunctionCall: *callback,
						Result:                     nil,
						Error:                      toolCall.Error,
					})
					if alter != nil {
						functionResponse = alter
					}
					if err != nil {
						return nil, err
					}
				}

				continue
			}

			// * invoke callback after execution with response
			if state.OnAfterFunctionCall != nil {
				alter, err := state.OnAfterFunctionCall(&CallbackAfterFunctionCall{
					CallbackBeforeFunctionCall: *callback,
					Result:                     functionResponse,
					Error:                      nil,
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
				if r.Option.ParseErrorBreak != nil && *r.Option.ParseErrorBreak {
					return nil, gut.Err(false, fmt.Sprintf("failed to marshal response for tool %s: %s", gut.Val(toolCall.Name), err.Error()), err)
				}
				toolCall.Error = gut.Ptr("failed to marshal response: " + err.Error())
				toolCalls = append(toolCalls, toolCall)
				continue
			}

			// * create tool result message
			toolCall.Result = responseJson
			toolCalls = append(toolCalls, toolCall)
		}

		toolMessage := &call.AssistantMessage{
			Content:   response.Message.Content,
			ToolCalls: toolCalls,
			Usage:     response.Message.Usage,
		}

		// * compact error messages
		if len(toolMessage.ToolCalls) == 1 &&
			len(state.ToolMessages) > 0 &&
			r.Option.ParseErrorCompact != nil &&
			*r.Option.ParseErrorCompact {
			for i := len(state.ToolMessages) - 1; i >= 0; i-- {
				tm := state.ToolMessages[i]
				if len(tm.ToolCalls) == 1 && tm.ToolCalls[0].Error != nil {
					if toolMessage.ToolCalls[0].Name == tm.ToolCalls[0].Name {
						// * remove previous error message
						state.ToolMessages = append(state.ToolMessages[:i], state.ToolMessages[i+1:]...)
					}
					break
				}
			}
		}

		// * call callback
		if state.OnToolMessage != nil {
			if err := state.OnToolMessage(toolMessage); err != nil {
				return nil, err
			}
		}

		// * append tool message to state
		state.ToolMessages = append(state.ToolMessages, toolMessage)
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
			InputSchema: declaration.ArgumentsSchema,
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

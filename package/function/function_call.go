package function

import (
	"encoding/json"

	"github.com/bsthun/gut"
	"go.scnd.dev/open/model-agentic/package/call"
)

type Caller interface {
	Call(request *Request, option *call.Option, output any, callback func(invoke *CallbackInvoke)) (*call.Response, *gut.ErrorInstance)
	AddDeclaration(declaration *Declaration)
}

type Call struct {
	Caller       call.Caller    `json:"-"`
	Declarations []*Declaration `json:"declarations"`
}

func New(caller call.Caller) Caller {
	return &Call{
		Caller:       caller,
		Declarations: []*Declaration{},
	}
}

func (c *Call) AddDeclaration(declaration *Declaration) {
	c.Declarations = append(c.Declarations, declaration)
}

func (c *Call) Call(request *Request, option *call.Option, output any, callback func(invoke *CallbackInvoke)) (*call.Response, *gut.ErrorInstance) {
	// * convert function request to call request by appending function declarations as tools
	callRequest := &call.Request{
		Model:       request.Model,
		Messages:    request.Messages,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		TopK:        request.TopK,
		Tools:       c.DeclarationsToTools(),
	}

	// * loop until no more tool calls
	for {
		// * call underlying caller
		response, err := c.Caller.Call(callRequest, option, output)
		if err != nil {
			return nil, err
		}

		// * check if there are tool calls
		if response.FinishReason != "tool_calls" || len(response.Message.ToolCalls) == 0 {
			return response, nil
		}

		// * process each tool call
		toolCalls := make([]*call.ToolCall, 0)
		for _, toolCall := range response.Message.ToolCalls {
			// * find matching declaration
			declaration := c.GetDeclaration(toolCall.Name)
			if declaration == nil {
				return nil, gut.Err(false, "declaration not found for tool: "+gut.Val(toolCall.Name), nil)
			}

			// * unmarshal arguments from json
			var arguments map[string]any
			if err := json.Unmarshal(toolCall.Arguments, &arguments); err != nil {
				return nil, gut.Err(false, "failed to unmarshal tool call arguments", err)
			}

			// * invoke callback before execution with response as nil
			if callback != nil {
				callback(&CallbackInvoke{
					ToolCallId:  toolCall.Id,
					Declaration: declaration,
					Argument:    arguments,
					Response:    nil,
				})
			}

			// * execute function to get response
			functionResponse, funcErr := declaration.Func(arguments)
			if funcErr != nil {
				return nil, funcErr
			}

			// * invoke callback after execution with response
			if callback != nil {
				callback(&CallbackInvoke{
					ToolCallId:  toolCall.Id,
					Declaration: declaration,
					Argument:    arguments,
					Response:    functionResponse,
				})
			}

			// * marshal response to json
			responseJSON, err := json.Marshal(functionResponse)
			if err != nil {
				return nil, gut.Err(false, "failed to marshal function response to json", err)
			}

			// * create tool result message
			toolCall.Result = responseJSON
			toolCalls = append(toolCalls, toolCall)
		}

		toolMessage := &call.Message{
			Role:      gut.Ptr("tool"),
			Content:   nil,
			Images:    nil,
			ToolCalls: toolCalls,
		}

		// * append tool result message
		callRequest.Messages = append(callRequest.Messages, toolMessage)
	}
}

func (c *Call) DeclarationsToTools() []*call.Tool {
	var tools []*call.Tool
	for _, declaration := range c.Declarations {
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

func (c *Call) GetDeclaration(name *string) *Declaration {
	if name == nil {
		return nil
	}
	for _, declaration := range c.Declarations {
		if declaration.Name != nil && *declaration.Name == *name {
			return declaration
		}
	}
	return nil
}

package call

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/bsthun/gut"
)

type Anthropic struct {
	Client *anthropic.Client
}

func NewAnthropic(baseUrl string, apiKey string) Caller {
	client := anthropic.NewClient(
		option.WithBaseURL(baseUrl),
		option.WithAPIKey(apiKey),
	)

	return &Anthropic{
		Client: &client,
	}
}

func (r *Anthropic) Call(request *Request, option *Option, output any) (*Response, *gut.ErrorInstance) {
	if request == nil || option == nil {
		return nil, gut.Err(false, "request or option is nil", nil)
	}

	// * convert request to anthropic message parameters
	messageParams := r.RequestToMessageParams(request, output)

	// * call anthropic api with retry logic
	maxRetries := 3
	var message *anthropic.Message
	var err error

	for i := 0; i < maxRetries; i++ {
		message, err = (*r.Client).Messages.New(context.Background(), messageParams)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			gut.Debug("anthropic retry %d due to error: %v", i+1, err)
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return nil, gut.Err(false, fmt.Sprintf("failed to call anthropic after %d retries", maxRetries), err)
	}

	// * convert anthropic response to internal format
	response := r.MessageToResponse(message, output)
	if response == nil {
		return nil, gut.Err(false, "invalid response from anthropic", nil)
	}

	return response, nil
}

func (r *Anthropic) RequestToMessageParams(request *Request, output any) anthropic.MessageNewParams {
	// * convert messages
	messages := r.RequestToMessages(request)

	// * build message parameters
	messageParams := anthropic.MessageNewParams{
		Messages: messages,
	}

	// * set model
	if request.Model != nil {
		messageParams.Model = anthropic.Model(*request.Model)
	}

	// * set max tokens (required for anthropic)
	if request.MaxTokens != nil {
		messageParams.MaxTokens = int64(*request.MaxTokens)
	} else {
		messageParams.MaxTokens = 4096 // default
	}

	// * set optional parameters
	if request.Temperature != nil {
		messageParams.Temperature = anthropic.Float(*request.Temperature)
	}
	if request.TopP != nil {
		messageParams.TopP = anthropic.Float(*request.TopP)
	}

	// * set tools if provided
	if len(request.Tools) > 0 {
		messageParams.Tools = r.RequestToTools(request.Tools)
	}

	// * set response format if output is provided
	if output != nil {
		// TODO: anthropic sdk not support response_format directly
	}

	return messageParams
}

func (r *Anthropic) RequestToMessages(request *Request) []anthropic.MessageParam {
	var messages []anthropic.MessageParam

	for _, msg := range request.Messages {
		if msg == nil || msg.Role == nil {
			continue
		}

		switch *msg.Role {
		case "system":
			// * system messages are handled at the top level in anthropic
			continue
		case "user":
			messages = append(messages, r.UserMessageToMessageParam(msg))
		case "assistant":
			messages = append(messages, r.AssistantMessageToMessageParam(msg))
		case "tool":
			// * tool results are added as user messages in anthropic
			for _, toolCall := range msg.ToolCalls {
				toolResultBlock := anthropic.NewToolResultBlock(*toolCall.Id, fmt.Sprintf("Name: %s, Request %s, Response: %s", *toolCall.Name, toolCall.Arguments, toolCall.Result), false)
				messages = append(messages, anthropic.NewUserMessage(toolResultBlock))
			}
		}
	}

	return messages
}

func (r *Anthropic) UserMessageToMessageParam(message *Message) anthropic.MessageParam {
	if message == nil {
		return anthropic.NewUserMessage(anthropic.NewTextBlock(""))
	}

	// * handle image content
	if len(message.Images) > 0 {
		var contentBlocks []anthropic.ContentBlockParamUnion

		// * add text content if present
		if message.Content != nil {
			contentBlocks = append(contentBlocks, anthropic.NewTextBlock(*message.Content))
		}

		// * add image content
		imageData := base64.StdEncoding.EncodeToString(message.Images)
		imageSource := anthropic.Base64ImageSourceParam{
			MediaType: anthropic.Base64ImageSourceMediaTypeImagePNG,
			Data:      imageData,
		}
		contentBlocks = append(contentBlocks, anthropic.NewImageBlock(imageSource))

		return anthropic.NewUserMessage(contentBlocks...)
	}

	// * handle text-only content
	content := ""
	if message.Content != nil {
		content = *message.Content
	}
	return anthropic.NewUserMessage(anthropic.NewTextBlock(content))
}

func (r *Anthropic) AssistantMessageToMessageParam(message *Message) anthropic.MessageParam {
	if message == nil {
		return anthropic.NewAssistantMessage(anthropic.NewTextBlock(""))
	}

	var contentBlocks []anthropic.ContentBlockParamUnion

	// * add text content if present
	if message.Content != nil && *message.Content != "" {
		contentBlocks = append(contentBlocks, anthropic.NewTextBlock(*message.Content))
	}

	// * add tool calls if present
	if len(message.ToolCalls) > 0 {
		for _, toolCall := range message.ToolCalls {
			if toolCall != nil {
				toolUseBlock := r.ToolCallToToolUseBlock(toolCall)
				contentBlocks = append(contentBlocks, toolUseBlock)
			}
		}
	}

	return anthropic.NewAssistantMessage(contentBlocks...)
}

func (r *Anthropic) ToolCallToToolUseBlock(toolCall *ToolCall) anthropic.ContentBlockParamUnion {
	name := ""
	input := make([]byte, 0)

	if toolCall.Name != nil {
		name = *toolCall.Name
	}
	if toolCall.Arguments != nil {
		input = toolCall.Arguments
	}

	toolUseID := ""
	if toolCall.Id != nil {
		toolUseID = *toolCall.Id
	}

	return anthropic.NewToolUseBlock(toolUseID, input, name)
}

func (r *Anthropic) MessageToResponse(message *anthropic.Message, output any) *Response {
	if message == nil || len(message.Content) == 0 {
		return nil
	}

	response := &Response{
		Id:           message.ID,
		Model:        string(message.Model),
		FinishReason: string(message.StopReason),
		Message:      r.MessageContentToMessage(message, output),
	}

	promptTokens := int(message.Usage.InputTokens)
	completionTokens := int(message.Usage.OutputTokens)
	response.Usage = &Usage{
		InputTokens:  &promptTokens,
		OutputTokens: &completionTokens,
	}

	return response
}

func (r *Anthropic) MessageContentToMessage(message *anthropic.Message, output any) *Message {
	role := "assistant"
	result := &Message{
		Role:      &role,
		Content:   nil,
		Images:    nil,
		ToolCalls: nil,
	}

	var content string
	var toolCalls []*ToolCall

	for _, contentBlock := range message.Content {
		switch contentBlock.Type {
		case "text":
			textBlock := contentBlock.AsText()
			if content != "" {
				content += "\n"
			}
			content += textBlock.Text
		case "tool_use":
			toolUseBlock := contentBlock.AsToolUse()
			toolCalls = append(toolCalls, r.ToolUseBlockToToolCall(toolUseBlock))
		}
	}

	// * set structured output
	if output != nil && content != "" {
		if err := json.Unmarshal([]byte(content), output); err != nil {
			gut.Debug("failed to unmarshal output content", err)
		}
	}

	// * set tool calls
	if len(toolCalls) > 0 {
		result.ToolCalls = toolCalls
	}

	return result
}

func (r *Anthropic) ToolUseBlockToToolCall(toolUseBlock anthropic.ToolUseBlock) *ToolCall {
	result := &ToolCall{
		Id:   &toolUseBlock.ID,
		Name: &toolUseBlock.Name,
	}

	if toolUseBlock.Input != nil {
		result.Arguments = toolUseBlock.Input
	}

	typeStr := "function"
	result.Type = &typeStr

	return result
}

func (r *Anthropic) RequestToTools(tools []*Tool) []anthropic.ToolUnionParam {
	var anthropicTools []anthropic.ToolUnionParam

	for _, tool := range tools {
		if tool == nil {
			continue
		}

		// * convert input schema
		var parameters anthropic.ToolInputSchemaParam
		if tool.InputSchema != nil {
			// * convert schema recursively to handle items properly
			schemaBytes, _ := json.Marshal(tool.InputSchema)
			_ = json.Unmarshal(schemaBytes, &parameters)
		}

		anthropicTool := anthropic.ToolUnionParamOfTool(parameters, "")

		// * set tool name
		if tool.Name != nil {
			anthropicTool = anthropic.ToolUnionParamOfTool(parameters, *tool.Name)
		}

		// * set tool description
		if tool.Description != nil {
			// * Note: Description is set through the tool construction
			// * The SDK doesn't seem to expose direct description setting in the union param
		}

		anthropicTools = append(anthropicTools, anthropicTool)
	}

	return anthropicTools
}

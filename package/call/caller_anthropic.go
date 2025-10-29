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

type ProviderAnthropic struct {
	Client *anthropic.Client
}

func NewAnthropic(baseUrl string, apiKey string) Caller {
	client := anthropic.NewClient(
		option.WithBaseURL(baseUrl),
		option.WithAPIKey(apiKey),
		option.WithAuthToken(apiKey),
	)

	return &ProviderAnthropic{
		Client: &client,
	}
}

func (r *ProviderAnthropic) Call(request *Request, option *Option, output any) (*Response, *gut.ErrorInstance) {
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

func (r *ProviderAnthropic) RequestToMessageParams(request *Request, output any) anthropic.MessageNewParams {
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

func (r *ProviderAnthropic) RequestToMessages(request *Request) []anthropic.MessageParam {
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
				toolResultBlock := anthropic.NewToolResultBlock(*toolCall.Id, toolCall.String(), false)
				messages = append(messages, anthropic.NewUserMessage(toolResultBlock))
			}
		}
	}

	return messages
}

func (r *ProviderAnthropic) UserMessageToMessageParam(message *Message) anthropic.MessageParam {
	if message == nil {
		return anthropic.NewUserMessage(anthropic.NewTextBlock(""))
	}

	// * handle image content
	if len(message.Image) > 0 || message.ImageUrl != nil {
		var contentBlocks []anthropic.ContentBlockParamUnion

		// * add text content if present
		if message.Content != nil {
			contentBlocks = append(contentBlocks, anthropic.NewTextBlock(*message.Content))
		}

		// * add image content
		var imageBlock anthropic.ContentBlockParamUnion
		if len(message.Image) > 0 {
			imageData := base64.StdEncoding.EncodeToString(message.Image)
			imageBlock = anthropic.NewImageBlock(anthropic.Base64ImageSourceParam{
				MediaType: anthropic.Base64ImageSourceMediaTypeImagePNG,
				Data:      imageData,
			})
		}
		if message.ImageUrl != nil {
			imageBlock = anthropic.NewImageBlock(anthropic.URLImageSourceParam{
				URL: *message.ImageUrl,
			})
		}
		contentBlocks = append(contentBlocks, imageBlock)

		return anthropic.NewUserMessage(contentBlocks...)
	}

	// * handle text-only content
	content := ""
	if message.Content != nil {
		content = *message.Content
	}
	return anthropic.NewUserMessage(anthropic.NewTextBlock(content))
}

func (r *ProviderAnthropic) AssistantMessageToMessageParam(message *Message) anthropic.MessageParam {
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

func (r *ProviderAnthropic) ToolCallToToolUseBlock(toolCall *ToolCall) anthropic.ContentBlockParamUnion {
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

func (r *ProviderAnthropic) MessageToResponse(message *anthropic.Message, output any) *Response {
	if message == nil || len(message.Content) == 0 {
		return nil
	}

	response := &Response{
		Id:           message.ID,
		Model:        string(message.Model),
		FinishReason: string(message.StopReason),
		Message:      r.MessageContentToMessage(message, output),
	}

	response.Usage = &Usage{
		InputTokens:  &message.Usage.InputTokens,
		OutputTokens: &message.Usage.OutputTokens,
		CachedTokens: gut.Ptr(message.Usage.CacheCreationInputTokens + message.Usage.CacheReadInputTokens),
	}

	return response
}

func (r *ProviderAnthropic) MessageContentToMessage(message *anthropic.Message, output any) *Message {
	role := "assistant"
	result := &Message{
		Role:        &role,
		Content:     nil,
		Image:       nil,
		ImageUrl:    nil,
		ImageDetail: nil,
		ToolCalls:   nil,
		Usage:       nil,
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

	// * unmarshal structured output
	if output != nil && content != "" {
		content = ContentClean(content)
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

func (r *ProviderAnthropic) ToolUseBlockToToolCall(toolUseBlock anthropic.ToolUseBlock) *ToolCall {
	result := &ToolCall{
		Id:   &toolUseBlock.ID,
		Name: &toolUseBlock.Name,
	}

	if toolUseBlock.Input != nil {
		result.Arguments = toolUseBlock.Input
	}

	result.Type = gut.Ptr("function")

	return result
}

func (r *ProviderAnthropic) RequestToTools(tools []*Tool) []anthropic.ToolUnionParam {
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

		// * set tool name
		// TODO: append tool description for anthropic if supported
		anthropicTool := anthropic.ToolUnionParamOfTool(parameters, *tool.Name)
		anthropicTools = append(anthropicTools, anthropicTool)
	}

	return anthropicTools
}

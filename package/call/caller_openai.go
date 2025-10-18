package call

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bsthun/gut"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

type Openai struct {
	Client *openai.Client
}

func NewOpenai(baseUrl string, apiKey string) Caller {
	client := openai.NewClient(
		option.WithBaseURL(baseUrl),
		option.WithAPIKey(apiKey),
	)

	return &Openai{
		Client: &client,
	}
}

func (r *Openai) Call(request *Request, option *Option, output any) (*Response, *gut.ErrorInstance) {
	if request == nil || option == nil {
		return nil, gut.Err(false, "request or option is nil", nil)
	}

	// * convert request to openai chat parameters
	chatParams := r.RequestToChatParams(request, option, output)

	// * call openai api with retry logic
	maxRetries := 3
	var chatCompletion *openai.ChatCompletion
	var err error

	for i := 0; i < maxRetries; i++ {
		chatCompletion, err = (*r.Client).Chat.Completions.New(context.Background(), chatParams)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			gut.Debug("openai retry %d due to error: %v", i+1, err)
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return nil, gut.Err(false, fmt.Sprintf("failed to call openai after %d retries", maxRetries), err)
	}

	// * convert openai response to internal format
	response := r.ChatCompletionToResponse(chatCompletion)
	if response == nil {
		return nil, gut.Err(false, "invalid response from openai", nil)
	}

	// * parse response content
	if output != nil && response.Message != nil && response.Message.Content != nil {
		if err := json.Unmarshal([]byte(*response.Message.Content), output); err != nil {
			return nil, gut.Err(false, "failed to unmarshal response content to output", err)
		}
	}

	return response, nil
}

func (r *Openai) RequestToChatParams(request *Request, option *Option, output any) openai.ChatCompletionNewParams {
	// * convert messages
	messages := r.RequestToMessages(request)

	// * build chat completion parameters
	chatParams := openai.ChatCompletionNewParams{
		Messages: messages,
	}

	// * set model
	if request.Model != nil {
		chatParams.Model = *request.Model
	}

	// * set optional parameters
	if request.MaxTokens != nil {
		chatParams.MaxCompletionTokens = openai.Int(int64(*request.MaxTokens))
	}
	if request.Temperature != nil {
		chatParams.Temperature = openai.Float(*request.Temperature)
	}
	if request.TopP != nil {
		chatParams.TopP = openai.Float(*request.TopP)
	}

	// * set tools if provided
	if len(request.Tools) > 0 {
		chatParams.Tools = r.RequestToTools(request.Tools)
	}

	// * set output format if output schema is provided
	if output != nil {
		schema := SchemaConvert(output)
		chatParams.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				Type: "json_schema",
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        gut.Val(option.SchemaName),
					Description: openai.String(gut.Val(option.SchemaDescription)),
					Schema:      schema,
					Strict:      openai.Bool(true),
				},
			},
		}
	}

	return chatParams
}

func (r *Openai) RequestToMessages(request *Request) []openai.ChatCompletionMessageParamUnion {
	var messages []openai.ChatCompletionMessageParamUnion

	for _, msg := range request.Messages {
		if msg == nil || msg.Role == nil {
			continue
		}

		switch *msg.Role {
		case "system":
			if msg.Content != nil {
				messages = append(messages, openai.SystemMessage(*msg.Content))
			}
		case "user":
			messages = append(messages, r.UserMessageToChatParam(msg))
		case "assistant":
			messages = append(messages, r.AssistantMessageToChatParam(msg))
		case "tool":
			for _, toolCall := range msg.ToolCalls {
				messages = append(messages, openai.ToolMessage(fmt.Sprintf("Name: %s, Request: %s, Response: %s", *toolCall.Name, toolCall.Arguments, toolCall.Result), *toolCall.Id))
			}
		}
	}

	return messages
}

func (r *Openai) UserMessageToChatParam(message *Message) openai.ChatCompletionMessageParamUnion {
	if message == nil {
		return openai.UserMessage("")
	}

	// * handle image content
	if len(message.Images) > 0 {
		var contentParts []openai.ChatCompletionContentPartUnionParam

		// * add text content if present
		if message.Content != nil {
			contentParts = append(contentParts, openai.TextContentPart(*message.Content))
		}

		// * add image content
		imageData := base64.StdEncoding.EncodeToString(message.Images)
		imageUrl := fmt.Sprintf("data:image/png;base64,%s", imageData)
		contentParts = append(contentParts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL: imageUrl,
		}))

		return openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Role: "user",
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfArrayOfContentParts: contentParts,
				},
			},
		}
	}

	// * handle text-only content
	content := ""
	if message.Content != nil {
		content = *message.Content
	}
	return openai.UserMessage(content)
}

func (r *Openai) AssistantMessageToChatParam(message *Message) openai.ChatCompletionMessageParamUnion {
	if message == nil {
		return openai.AssistantMessage("")
	}

	content := ""
	if message.Content != nil {
		content = *message.Content
	}

	return openai.AssistantMessage(content)
}

func (r *Openai) ChatCompletionToResponse(completion *openai.ChatCompletion) *Response {
	if completion == nil || len(completion.Choices) == 0 {
		return nil
	}

	choice := completion.Choices[0]
	response := &Response{
		Id:           completion.ID,
		Model:        completion.Model,
		FinishReason: choice.FinishReason,
		Message:      r.ChatCompletionMessageToMessage(choice.Message),
	}

	promptTokens := int(completion.Usage.PromptTokens)
	completionTokens := int(completion.Usage.CompletionTokens)
	response.Usage = &Usage{
		InputTokens:  &promptTokens,
		OutputTokens: &completionTokens,
	}

	return response
}

func (r *Openai) ChatCompletionMessageToMessage(message openai.ChatCompletionMessage) *Message {
	result := new(Message)

	if message.Role != "" {
		role := string(message.Role)
		result.Role = &role
	}

	if message.Content != "" {
		result.Content = &message.Content
	}

	if len(message.ToolCalls) > 0 {
		var toolCalls []*ToolCall
		for _, toolCall := range message.ToolCalls {
			toolCalls = append(toolCalls, r.ChatCompletionToolCallToToolCall(toolCall))
		}
		result.ToolCalls = toolCalls
	}

	// * handle reasoning conten
	if message.Content != "" {
		result.Content = &message.Content
	}

	return result
}

func (r *Openai) RequestToTools(tools []*Tool) []openai.ChatCompletionToolParam {
	var openaiTools []openai.ChatCompletionToolParam

	for _, tool := range tools {
		if tool == nil {
			continue
		}

		// * convert input schema with proper items handling
		var parameters openai.FunctionParameters
		if tool.InputSchema != nil {
			schemaBytes, _ := json.Marshal(tool.InputSchema)
			_ = json.Unmarshal(schemaBytes, &parameters)
		}

		openaiTool := openai.ChatCompletionToolParam{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        "",
				Description: openai.String(""),
				Parameters:  parameters,
			},
		}

		// * set tool name
		if tool.Name != nil {
			openaiTool.Function.Name = *tool.Name
		}

		// * set tool description
		if tool.Description != nil {
			openaiTool.Function.Description = openai.String(*tool.Description)
		}

		openaiTools = append(openaiTools, openaiTool)
	}

	return openaiTools
}

func (r *Openai) ChatCompletionToolCallToToolCall(toolCall openai.ChatCompletionMessageToolCall) *ToolCall {
	typeStr := string(toolCall.Type)
	result := &ToolCall{
		Id:        &toolCall.ID,
		Type:      &typeStr,
		Name:      nil,
		Arguments: nil,
		Result:    nil,
	}

	if toolCall.Function.Name != "" {
		result.Name = &toolCall.Function.Name
		result.Arguments = []byte(toolCall.Function.Arguments)
	}

	return result
}

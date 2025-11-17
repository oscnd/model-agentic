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
	"github.com/openai/openai-go/packages/respjson"
	"github.com/openai/openai-go/shared"
)

type ProviderOpenai struct {
	Client *openai.Client
}

func NewOpenai(baseUrl string, apiKey string) Caller {
	client := openai.NewClient(
		option.WithBaseURL(baseUrl),
		option.WithAPIKey(apiKey),
	)

	return &ProviderOpenai{
		Client: &client,
	}
}

func (r *ProviderOpenai) Call(request *Request, option *Option, output any) (*Response, *gut.ErrorInstance) {
	if request == nil || option == nil {
		return nil, gut.Err(false, "request or option is nil", nil)
	}

	// * convert request to openai chat parameters
	chatParams := r.RequestToChatParams(request, option, output)

	// * initialize completion struct
	completion := &openai.ChatCompletion{
		ID:      "completion-" + *gut.Random(gut.RandomSet.MixedAlphaNum, 12),
		Model:   chatParams.Model,
		Created: time.Now().Unix(),
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: "",
				},
				FinishReason: "",
			},
		},
		Usage: openai.CompletionUsage{},
	}

	// * call openai streaming api
	stream := r.Client.Chat.Completions.NewStreaming(context.Background(), chatParams)
	for stream.Next() {
		chunk := stream.Current()

		// * update model
		if chunk.Model != "" {
			completion.Model = chunk.Model
		}

		// * update usage
		if chunk.Usage.PromptTokens > 0 || chunk.Usage.CompletionTokens > 0 {
			completion.Usage = chunk.Usage
		}

		// * process choices
		for _, choice := range chunk.Choices {
			// * find existing choice by index
			for int64(len(completion.Choices)) <= choice.Index {
				completion.Choices = append(completion.Choices, openai.ChatCompletionChoice{
					Index: int64(len(completion.Choices)),
					Message: openai.ChatCompletionMessage{
						Role: "assistant",
					},
				})
			}

			// * accumulate content
			if choice.Delta.Content != "" {
				completion.Choices[choice.Index].Message.Content += choice.Delta.Content
			}

			// * accumulate tool calls
			for _, toolCallDelta := range choice.Delta.ToolCalls {
				toolCallIndex := int(toolCallDelta.Index)

				// * ensure toolCalls array has enough elements
				for len(completion.Choices[choice.Index].Message.ToolCalls) <= toolCallIndex {
					completion.Choices[choice.Index].Message.ToolCalls = append(completion.Choices[choice.Index].Message.ToolCalls, openai.ChatCompletionMessageToolCall{
						Type:     "function",
						Function: openai.ChatCompletionMessageToolCallFunction{},
					})
				}

				// * update tool call id
				if toolCallDelta.ID != "" {
					completion.Choices[choice.Index].Message.ToolCalls[toolCallIndex].ID = toolCallDelta.ID
				}

				// * update function name
				if toolCallDelta.Function.Name != "" {
					completion.Choices[choice.Index].Message.ToolCalls[toolCallIndex].Function.Name = toolCallDelta.Function.Name
				}

				// * accumulate function arguments
				if toolCallDelta.Function.Arguments != "" {
					completion.Choices[choice.Index].Message.ToolCalls[toolCallIndex].Function.Arguments += toolCallDelta.Function.Arguments
				}
			}

			// * capture finish reason
			if choice.FinishReason != "" {
				completion.Choices[choice.Index].FinishReason = choice.FinishReason
			}

			// * accumulate extra fields
			if choice.Delta.JSON.ExtraFields != nil {
				if completion.Choices[choice.Index].JSON.ExtraFields == nil {
					completion.Choices[choice.Index].JSON.ExtraFields = make(map[string]respjson.Field)
				}
				for k, v := range choice.Delta.JSON.ExtraFields {
					var val string
					if err := json.Unmarshal([]byte(v.Raw()), &val); err != nil {
						continue
					}
					val = completion.Choices[choice.Index].JSON.ExtraFields[k].Raw() + val
					completion.Choices[choice.Index].JSON.ExtraFields[k] = respjson.NewField(val)
				}
			}
		}

		if option.OnResponse != nil {
			response := r.ChatCompletionToResponse(completion)
			option.OnResponse(response)
		}
	}

	// * check for streaming errors
	if err := stream.Err(); err != nil {
		return nil, gut.Err(false, fmt.Sprintf("openai streaming failed: %s", err), err)
	}

	// * convert completion to internal format
	response := r.ChatCompletionToResponse(completion)
	if response == nil {
		return nil, gut.Err(false, "invalid response from openai", nil)
	}

	// * parse response content
	if output != nil && response.Message != nil && response.Message.Content != nil {
		*response.Message.Content = ContentClean(*response.Message.Content)
		if err := json.Unmarshal([]byte(*response.Message.Content), output); err != nil {
			return nil, gut.Err(false, "failed to unmarshal response content to output", err)
		}
	}

	return response, nil
}

func (r *ProviderOpenai) RequestToChatParams(request *Request, option *Option, output any) openai.ChatCompletionNewParams {
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

	// * set reasoning effort if provided
	if request.ReasoningEffort != nil {
		chatParams.ReasoningEffort = shared.ReasoningEffort(*request.ReasoningEffort)
	}

	// * set tools if provided
	if len(request.Tools) > 0 {
		chatParams.ParallelToolCalls = openai.Bool(true)
		chatParams.Tools = r.RequestToTools(request.Tools)
	}

	// * set extra fields from option
	if request.ExtraFields != nil {
		chatParams.SetExtraFields(request.ExtraFields)
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

func (r *ProviderOpenai) RequestToMessages(request *Request) []openai.ChatCompletionMessageParamUnion {
	var messages []openai.ChatCompletionMessageParamUnion

	for _, message := range request.Messages {
		if message == nil {
			continue
		}

		switch message.(type) {
		case *SystemMessage:
			m := message.(*SystemMessage)
			if m.Content != nil {
				messages = append(messages, openai.SystemMessage(*m.Content))
			}
		case *UserMessage:
			m := message.(*UserMessage)
			messages = append(messages, r.UserMessageToChatParam(m))
		case *AssistantMessage:
			m := message.(*AssistantMessage)
			ok, mm := r.AssistantMessageToChatParam(m)
			if ok {
				messages = append(messages, mm)
			}
			for _, toolCall := range m.ToolCalls {
				messages = append(messages, openai.ToolMessage(toolCall.String(), *toolCall.Id))
			}
		}
	}

	return messages
}

func (r *ProviderOpenai) UserMessageToChatParam(message *UserMessage) openai.ChatCompletionMessageParamUnion {
	if message == nil {
		return openai.UserMessage("")
	}

	// * handle image content
	if len(message.Image) > 0 || message.ImageUrl != nil {
		var contentParts []openai.ChatCompletionContentPartUnionParam

		// * add text content if present
		if message.Content != nil {
			contentParts = append(contentParts, openai.TextContentPart(*message.Content))
		}

		// * construct image url
		var imageUrl string
		if len(message.Image) > 0 {
			imageData := base64.StdEncoding.EncodeToString(message.Image)
			imageUrl = fmt.Sprintf("data:image/png;base64,%s", imageData)
		}
		if message.ImageUrl != nil {
			imageUrl = *message.ImageUrl
		}
		contentParts = append(contentParts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL:    imageUrl,
			Detail: gut.Val(message.ImageDetail),
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

func (r *ProviderOpenai) AssistantMessageToChatParam(message *AssistantMessage) (bool, openai.ChatCompletionMessageParamUnion) {
	if message == nil {
		return false, openai.ChatCompletionMessageParamUnion{}
	}

	if message.Content == nil || *message.Content == "" {
		return false, openai.ChatCompletionMessageParamUnion{}
	}

	return true, openai.AssistantMessage(*message.Content)
}

func (r *ProviderOpenai) ChatCompletionToResponse(completion *openai.ChatCompletion) *Response {
	if completion == nil || len(completion.Choices) == 0 {
		return nil
	}

	choice := completion.Choices[0]
	response := &Response{
		Id:           completion.ID,
		Model:        completion.Model,
		FinishReason: choice.FinishReason,
		Message:      r.ChatCompletionMessageToMessage(choice.Message),
		TotalUsage:   nil,
		ExtraFields:  nil,
	}

	response.Message.Usage = &Usage{
		InputTokens:  &completion.Usage.PromptTokens,
		OutputTokens: &completion.Usage.CompletionTokens,
		CachedTokens: &completion.Usage.PromptTokensDetails.CachedTokens,
	}

	if choice.JSON.ExtraFields != nil {
		response.ExtraFields = make(map[string]string)
		for k, v := range choice.JSON.ExtraFields {
			if v.Valid() {
				response.ExtraFields[k] = v.Raw()
			}
		}
	}

	return response
}

func (r *ProviderOpenai) ChatCompletionMessageToMessage(message openai.ChatCompletionMessage) *AssistantMessage {
	result := new(AssistantMessage)

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

	// * handle reasoning content
	if message.Content != "" {
		result.Content = &message.Content
	}

	return result
}

func (r *ProviderOpenai) RequestToTools(tools []*Tool) []openai.ChatCompletionToolParam {
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

func (r *ProviderOpenai) ChatCompletionToolCallToToolCall(toolCall openai.ChatCompletionMessageToolCall) *ToolCall {
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

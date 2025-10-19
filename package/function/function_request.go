package function

import "go.scnd.dev/open/model/agentic/package/call"

type Request struct {
	Model       *string         `json:"model,omitempty"`
	Messages    []*call.Message `json:"messages,omitempty"`
	MaxTokens   *int            `json:"maxTokens,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	TopP        *float64        `json:"topP,omitempty"`
	TopK        *int            `json:"topK,omitempty"`
}

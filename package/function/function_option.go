package function

import "go.scnd.dev/open/model/agentic/package/call"

type Option struct {
	Model       *string      `json:"model,omitempty"`
	MaxTokens   *int         `json:"maxTokens,omitempty"`
	Temperature *float64     `json:"temperature,omitempty"`
	TopP        *float64     `json:"topP,omitempty"`
	TopK        *int         `json:"topK,omitempty"`
	CallOption  *call.Option `json:"callOption,omitempty"`
}

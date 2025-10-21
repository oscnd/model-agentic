package function

import "go.scnd.dev/open/model/agentic/package/call"

type Option struct {
	Model       *string      `json:"model"`
	MaxTokens   *int         `json:"maxTokens"`
	Temperature *float64     `json:"temperature"`
	TopP        *float64     `json:"topP"`
	TopK        *int         `json:"topK"`
	CallOption  *call.Option `json:"callOption"`
}

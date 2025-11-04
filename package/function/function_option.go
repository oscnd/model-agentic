package function

import "go.scnd.dev/open/model/agentic/package/call"

// Option contains configuration for function calling, extended call options.
type Option struct {
	Model             *string               `json:"model"`
	MaxTokens         *int                  `json:"maxTokens"`
	Temperature       *float64              `json:"temperature"`
	TopP              *float64              `json:"topP"`
	TopK              *int                  `json:"topK"`
	ReasoningEffort   *call.ReasoningEffort `json:"reasoningEffort"`
	ParseErrorBreak   *bool                 `json:"parseErrorBreak"`
	ParseErrorCompact *bool                 `json:"parseErrorTruncate"`
	CallOption        *call.Option          `json:"callOption"`
}

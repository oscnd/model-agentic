package call

// Request represents the unified request to be sent to a language model or agent,
// and it will be translated to provider-specific request format internally
type Request struct {
	Model       *string    `json:"model,omitempty"`
	MaxTokens   *int       `json:"maxTokens,omitempty"`
	Temperature *float64   `json:"temperature,omitempty"`
	TopP        *float64   `json:"topP,omitempty"`
	TopK        *int       `json:"topK,omitempty"`
	Messages    []*Message `json:"messages,omitempty"`
	Tools       []*Tool    `json:"tools,omitempty"`
}

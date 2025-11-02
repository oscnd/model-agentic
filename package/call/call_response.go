package call

// Response represents the response from a call to a language model or agent
type Response struct {
	Id           string            `json:"id,omitempty"`
	Model        string            `json:"model"`
	FinishReason string            `json:"finishReason"`
	Message      *AssistantMessage `json:"message"`
	TotalUsage   *Usage            `json:"totalUsage,omitempty"`
}

// Usage represents token usage information in a call response
type Usage struct {
	InputTokens  *int64 `json:"inputTokens,omitempty"`
	OutputTokens *int64 `json:"outputTokens,omitempty"`
	CachedTokens *int64 `json:"cachedTokens,omitempty"`
}

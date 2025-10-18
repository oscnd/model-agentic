package call

type Response struct {
	Id           string   `json:"id,omitempty"`
	Model        string   `json:"model"`
	FinishReason string   `json:"finishReason"`
	Message      *Message `json:"message"`
	Usage        *Usage   `json:"usage,omitempty"`
}

type Usage struct {
	InputTokens  *int `json:"inputTokens,omitempty"`
	OutputTokens *int `json:"outputTokens,omitempty"`
	CachedTokens *int `json:"cachedTokens,omitempty"`
}

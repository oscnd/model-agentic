package call

// Message represents a message with a language model or agent,
// if it used in request, it will be a conversation history with resulted tool calls,
// if it used in response, it will be the model's response with a pending execution (non-resulted) tool calls
type Message struct {
	Role      *string     `json:"role"`
	Content   *string     `json:"content"`
	Images    []byte      `json:"images,omitempty"`
	ToolCalls []*ToolCall `json:"toolCalls,omitempty"`
	Usage     *Usage      `json:"usage,omitempty"`
}

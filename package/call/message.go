package call

type Message struct {
	Role       string      `json:"role"`
	Content    string      `json:"content"`
	Images     []byte      `json:"images,omitempty"`
	ToolCallId *string     `json:"toolCallId,omitempty"`
	ToolCalls  []*ToolCall `json:"toolCalls,omitempty"`
}

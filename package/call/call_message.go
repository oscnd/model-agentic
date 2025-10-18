package call

type Message struct {
	Role      *string     `json:"role"`
	Content   *string     `json:"content"`
	Images    []byte      `json:"images,omitempty"`
	ToolCalls []*ToolCall `json:"toolCalls,omitempty"`
}

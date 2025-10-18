package function

type CallbackInvoke struct {
	ToolCallId  *string        `json:"toolCallId"`
	Declaration *Declaration   `json:"declaration"`
	Argument    map[string]any `json:"argument"`
	Response    map[string]any `json:"response"`
}

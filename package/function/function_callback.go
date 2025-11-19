package function

type CallbackBeforeFunctionCall struct {
	ToolCallId  *string      `json:"toolCallId"`
	Declaration *Declaration `json:"declaration"`
	Arguments   any          `json:"arguments"`
}

type CallbackAfterFunctionCall struct {
	CallbackBeforeFunctionCall
	Result map[string]any `json:"result"`
	Error  *string        `json:"error,omitempty"`
}

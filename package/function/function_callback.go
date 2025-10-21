package function

type CallbackBeforeFunctionCall struct {
	ToolCallId  *string        `json:"toolCallId"`
	Declaration *Declaration   `json:"declaration"`
	Argument    map[string]any `json:"argument"`
}

type CallbackAfterFunctionCall struct {
	CallbackBeforeFunctionCall
	Result map[string]any `json:"result"`
}

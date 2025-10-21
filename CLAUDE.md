# Claude

General guideline:

- Use pointer for struct.
- Use r as the receiver name. Example: `func (r *Handler) HandleOrganizationCreate(c *fiber.Ctx) error`.
- When initializing blank struct, use `x := new(Type)` instead of `x := &Type{}`.

## Commenting Style

### Style Guidelines

- Comment in format of `// * lowercase compact action` for each step, the comment should be in present tense without any
  additional explanation. Except comment for godoc.
- Comments should be present tense and directly describe the action
- No additional explanations in step comments - keep them compact
- Package comments should describe what the package provides and handles

### Examples

```go
// Package function provides function calling capabilities with state management,
// it handles function declarations, looped function calling execution, and callback hooks.
package function

// Option contains configuration for function calling, extended call options.
type Option struct {
	Model *string `json:"model"`
	// ...
}

// Call executes the function calling loop with state management and callbacks
// during execution, state passed can use to get current messages and set callbacks
func (r *Call) Call(state *State, output any) (*call.Response, *gut.ErrorInstance) {
	// * convert function request to call request
	callRequest := &call.Request{...}

	// * loop until no more tool calls
	for {
		// * call underlying caller
		response, err := r.Caller.Call(callRequest, r.Option.CallOption, output)
		// ...
	}
}
```
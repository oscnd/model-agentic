// Package call provides a unified interface for making language model inference calls,
// it handles schematic outputs and inference service providers.
package call

import "github.com/bsthun/gut"

// Caller defines the interface for making inference calls
type Caller interface {
	// Call executes a request to the language model inference service
	// request: Request containing messages, tools, model and other configurable parameters
	// option: Additional options for schema handling
	// output: The target structure to unmarshal the response into
	// Returns the model response and any error encountered
	Call(request *Request, option *Option, output any) (*Response, *gut.ErrorInstance)
}

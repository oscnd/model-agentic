// Package call provides a unified interface for making language model inference calls,
// it handles schematic outputs and inference service providers.
package call

import "github.com/bsthun/gut"

// Caller defines the interface for making inference calls
type Caller interface {
	// Call executes a request to the language model inference service
	// output: a struct to parse structured output of response content into
	Call(request *Request, option *Option, output any) (*Response, *gut.ErrorInstance)
}

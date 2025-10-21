package function

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
)

// DeclarationFunc defines the function signature for function implementations
type DeclarationFunc func(args map[string]any) (map[string]any, *gut.ErrorInstance)

// Declaration represents a function declaration with metadata and implementation
type Declaration struct {
	Name        *string         `json:"name"`
	Description *string         `json:"description"`
	Argument    *call.Schema    `json:"parameters"`
	Func        DeclarationFunc `json:"-"`
}

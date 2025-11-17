package function

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
)

// DeclarationFunc defines the function signature for function implementations
type DeclarationFunc func(arguments any) (map[string]any, *gut.ErrorInstance)

// Declaration represents a function declaration with metadata and implementation
type Declaration struct {
	Name            *string         `json:"name"`
	Description     *string         `json:"description"`
	Source          *string         `json:"source"`
	Arguments       any             `json:"arguments"`
	ArgumentsSchema *call.Schema    `json:"-"`
	Func            DeclarationFunc `json:"-"`
}

func NewDeclaration[T any](
	name *string,
	description *string,
	function func(arguments *T) (map[string]any, *gut.ErrorInstance),
) *Declaration {
	return &Declaration{
		Name:            name,
		Description:     description,
		Source:          nil,
		Arguments:       new(T),
		ArgumentsSchema: call.SchemaConvert(new(T)),
		Func: func(arguments any) (map[string]any, *gut.ErrorInstance) {
			if arguments == nil {
				return function(new(T))
			}

			parsed, ok := arguments.(*T)
			if !ok {
				return nil, gut.Err(false, "invalid argument type")
			}

			return function(parsed)
		},
	}
}

package function

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
)

type DeclarationFunc func(args map[string]any) (map[string]any, *gut.ErrorInstance)

type Declaration struct {
	Name        *string         `json:"name"`
	Description *string         `json:"description"`
	Argument    *call.Schema    `json:"parameters"`
	Func        DeclarationFunc `json:"-"`
}

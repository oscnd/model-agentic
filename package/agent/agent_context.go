package agent

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
)

// ContextPush pushes additional context to the agent's message history
func (r *Agent) ContextPush(content string) {
	r.Messages = append(r.Messages, &call.Message{
		Role:        gut.Ptr(call.RoleSystem),
		Content:     gut.Ptr("Additional context: " + content),
		Image:       nil,
		ImageUrl:    nil,
		ImageDetail: nil,
		ToolCalls:   nil,
		Usage:       nil,
	})
}

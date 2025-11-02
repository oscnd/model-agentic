package agent

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
)

// ContextPush pushes additional context to the agent's message history
func (r *Agent) ContextPush(content string) {
	r.Messages = append(r.Messages, &call.SystemMessage{
		Content: gut.Ptr("Additional context: " + content),
	})
}

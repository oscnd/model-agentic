package agent

import (
	"github.com/bsthun/gut"
	"go.scnd.dev/open/model/agentic/package/call"
)

func (r *Agent) ContextPush(content string) {
	r.Messages = append(r.Messages, &call.Message{
		Role:      gut.Ptr("system"),
		Content:   gut.Ptr("Additional context: " + content),
		Images:    nil,
		ToolCalls: nil,
	})
}

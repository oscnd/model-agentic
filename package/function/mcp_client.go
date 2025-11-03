package function

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bsthun/gut"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// McpClient wraps an MCP client for tool execution
type McpClient struct {
	client   client.MCPClient
	toolName *string
}

// NewMcpClient creates a new MCP client wrapper
func NewMcpClient(r client.MCPClient, toolName *string) *McpClient {
	return &McpClient{
		client:   r,
		toolName: toolName,
	}
}

// Execute calls the MCP tool with the provided arguments
func (r *McpClient) Execute(args map[string]any) (map[string]any, *gut.ErrorInstance) {
	ctx := context.Background()

	// * create mcp call tool request
	callRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      *r.toolName,
			Arguments: args,
		},
	}

	// * execute tool via mcp
	toolResult, err := r.client.CallTool(ctx, callRequest)
	if err != nil {
		return nil, gut.Err(false, fmt.Sprintf("failed to call mcp tool: %v", err))
	}

	// * convert tool result to response
	var result map[string]any
	if len(toolResult.Content) > 0 {
		// * select content
		content := toolResult.Content[0]

		if textContent, ok := mcp.AsTextContent(content); ok {
			// * fallback for empty result
			if len(textContent.Text) == 0 {
				return map[string]any{
					"success": true,
				}, nil
			}

			// * unmarshal text content
			if err := json.Unmarshal([]byte(textContent.Text), &result); err != nil {
				// * fallback to raw text
				result = map[string]any{
					"r": textContent.Text,
				}
			}
		} else {
			return nil, gut.Err(false, "unsupported mcp content type", nil)
		}
	}

	return result, nil
}

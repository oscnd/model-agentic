package function

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bsthun/gut"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type McpClient struct {
	client   client.MCPClient
	toolName *string
}

func NewMcpClient(r client.MCPClient, toolName *string) *McpClient {
	return &McpClient{
		client:   r,
		toolName: toolName,
	}
}

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
	var result []byte
	if len(toolResult.Content) > 0 {
		// * handle mcp content types
		content := toolResult.Content[0]
		if textContent, ok := mcp.AsTextContent(content); ok {
			result = []byte(textContent.Text)
		} else {
			// * fallback to json marshal
			bytes, err := json.Marshal(content)
			if err != nil {
				return nil, gut.Err(false, fmt.Sprintf("failed to marshal content: %v", err))
			}
			result = bytes
		}
	}

	// * return result as map
	if len(result) == 0 {
		return map[string]any{
			"success": true,
		}, nil
	}

	var resultMap map[string]any
	err = json.Unmarshal(result, &resultMap)
	if err != nil {
		return nil, gut.Err(false, fmt.Sprintf("failed to unmarshal tool result: %v", err))
	}

	return resultMap, nil
}

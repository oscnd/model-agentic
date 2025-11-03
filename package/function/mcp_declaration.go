package function

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"go.scnd.dev/open/model/agentic/package/call"
)

type McpOption struct {
	BaseUrl    string
	Header     map[string]string
	HttpClient *http.Client
}

// McpDeclarations fetches function declarations from an MCP server
func McpDeclarations(option *McpOption) ([]*Declaration, error) {
	ctx := context.Background()

	// * validate option
	if option == nil || option.BaseUrl == "" {
		return nil, fmt.Errorf("mcp option or base url is empty")
	}

	// * create mcp client
	mcpOptions := make([]transport.StreamableHTTPCOption, 0)
	if option.Header != nil {
		mcpOptions = append(mcpOptions, transport.WithHTTPHeaders(option.Header))
	}
	if option.HttpClient != nil {
		mcpOptions = append(mcpOptions, transport.WithHTTPBasicClient(option.HttpClient))
	}
	mcpClient, err := client.NewStreamableHttpClient(option.BaseUrl, mcpOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create mcp client: %w", err)
	}

	// * start client connection
	err = mcpClient.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start mcp client: %w", err)
	}

	// * initialize client with server
	initRequest := mcp.InitializeRequest{
		Request: mcp.Request{},
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities: mcp.ClientCapabilities{
				Roots: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{
					ListChanged: false,
				},
				Sampling: &struct{}{},
			},
			ClientInfo: mcp.Implementation{
				Name:    "agentic",
				Version: "1.0.0",
			},
		},
		Header: nil,
	}

	_, err = mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mcp client: %w", err)
	}

	// * list available tools
	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := mcpClient.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	// * convert mcp tools to declarations
	var declarations []*Declaration
	for _, tool := range toolsResult.Tools {
		// * convert input schema to call.Schema
		schema, err := McpSchemaToCallSchema(tool.InputSchema)
		if err != nil {
			continue
		}

		wrapper := NewMcpClient(mcpClient, &tool.Name)

		declaration := &Declaration{
			Name:        &tool.Name,
			Description: &tool.Description,
			Argument:    schema,
			Func:        wrapper.Execute,
		}
		declarations = append(declarations, declaration)
	}

	return declarations, nil
}

// McpSchemaToCallSchema converts MCP tool input schema to call schema
func McpSchemaToCallSchema(inputSchema mcp.ToolInputSchema) (*call.Schema, error) {
	schema := new(call.Schema)

	// * set type
	if inputSchema.Type != "" {
		schema.Type = &inputSchema.Type
	}

	// * set required
	if len(inputSchema.Required) > 0 {
		schema.Required = make([]*string, len(inputSchema.Required))
		for i, req := range inputSchema.Required {
			reqCopy := req
			schema.Required[i] = &reqCopy
		}
	}

	// * set properties
	if len(inputSchema.Properties) > 0 {
		schema.Properties = make(map[string]*call.Schema)
		for key, value := range inputSchema.Properties {
			propSchema := call.SchemaConvert(value)
			schema.Properties[key] = propSchema
		}
	}

	return schema, nil
}

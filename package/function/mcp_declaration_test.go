package function

import (
	"os"
	"strings"
	"testing"

	"github.com/bsthun/gut"
	"github.com/stretchr/testify/assert"
	"go.scnd.dev/open/model/agentic/package/call"
)

func TestMcpDeclarations(t *testing.T) {
	t.Run("FetchDeclarationsFromMcpServer", func(t *testing.T) {
		// * fetch declarations from mcp server
		mcpUrl := "http://localhost:3300/mcp"
		declarations, err := McpDeclarations(mcpUrl)

		// * assert no error
		assert.Nil(t, err)
		assert.NotNil(t, declarations)
		assert.Greater(t, len(declarations), 0, "should have at least one declaration")

		// * verify declaration structure
		for _, declaration := range declarations {
			assert.NotNil(t, declaration.Name)
			assert.NotNil(t, declaration.Description)
			assert.NotNil(t, declaration.Argument)
			assert.NotNil(t, declaration.Func)
		}
	})

	t.Run("SearchForPackageNameUsingMcpFunction", func(t *testing.T) {
		// * create caller
		caller := call.NewOpenai(os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))
		model := os.Getenv("OPENAI_MODEL")
		functionCall := New(caller)

		// * fetch declarations from mcp server
		mcpUrl := "http://localhost:3300/mcp"
		declarations, err := McpDeclarations(mcpUrl)
		assert.Nil(t, err)
		assert.NotNil(t, declarations)
		assert.Greater(t, len(declarations), 0, "should have at least one declaration")

		// * add all mcp declarations to function call
		for _, declaration := range declarations {
			functionCall.AddDeclaration(declaration)
		}

		// * create request to search for package name
		maxTokens := 500
		temperature := 0.7
		request := &Request{
			Model:       &model,
			MaxTokens:   &maxTokens,
			Temperature: &temperature,
			Messages: []*call.Message{
				{
					Role:    gut.Ptr("user"),
					Content: gut.Ptr("Please search for the package name of https://pkg.go.dev/go.scnd.dev/open/model/agentic using the available functions. Tell me what the package name is."),
				},
			},
		}

		// * track invocations
		var invocations []*CallbackInvoke
		var finalResponse string

		// * call function
		response, callErr := functionCall.Call(request, new(call.Option), nil, func(invoke *CallbackInvoke) {
			invocations = append(invocations, invoke)
		})

		// * assert no error
		assert.Nil(t, callErr)
		assert.NotNil(t, response)

		// * get final response content
		if response.Message != nil && response.Message.Content != nil {
			finalResponse = *response.Message.Content
		}

		// * assert at least one function was called
		assert.Greater(t, len(invocations), 0, "at least one function should be called")

		// * verify response contains "agentic"
		assert.Contains(t, strings.ToLower(finalResponse), "agentic", "response should contain 'agentic'")

		// * log invocations for debugging
		t.Logf("Total invocations: %d", len(invocations))
		for i, inv := range invocations {
			if inv.Declaration != nil && inv.Declaration.Name != nil {
				t.Logf("Invocation %d: %s", i+1, *inv.Declaration.Name)
			}
		}
		t.Logf("Final response: %s", finalResponse)
	})
}

# agentic

Language model agentic framework

## Usage

Install package using go get:

```bash
go get go.scnd.dev/model/agentic
```

Initialize an agent with a language model and tools:

```go
package main

import (
	"os"

	"github.com/bsthun/gut"
	"github.com/davecgh/go-spew/spew"
	"go.scnd.dev/model/agentic/package/call"
)

func main() {
	caller := call.NewOpenai(os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))

	type Person struct {
		Name    string   `json:"name" description:"The name of the person" validate:"required"`
		Emails  []string `json:"email" description:"The email address" validate:"required,email"`
		Address *struct {
			Street string `json:"street" validate:"required"`
			City   string `json:"city" validate:"required"`
		} `json:"address" validate:"required"`
	}

	request := &call.Request{
		Model:     gut.Ptr(os.Getenv("OPENAI_MODEL")),
		MaxTokens: gut.Ptr(1024),
		Messages: []*call.Message{
			{
				Role:    gut.Ptr("user"),
				Content: gut.Ptr("Generate information about a person named John who is 30 years old, lives in Thai."),
			},
		},
	}

	option := &call.Option{
		SchemaName:        gut.Ptr("Person"),
		SchemaDescription: gut.Ptr("Person information with 2 email addresses"),
	}

	output := new(Person)

	response, err := caller.Call(request, option, output)
	if err != nil {
		panic(err)
	}

	spew.Dump(response, output)
}

```

## Test

To run the tests, use the following command:

```bash
go test ./...
```

In order to pass some integration tests, these environment variables need to be set:

- `OPENAI_BASE_URL`: The base URL for OpenAI-compatible inference service.
- `OPENAI_API_KEY`: The API key for accessing the OpenAI-compatible inference service.
- `OPENAI_MODEL`: The specific OpenAI model to be used during testing.
- `ANTHROPIC_BASE_URL`: The base URL for Anthropic-compatible inference service.
- `ANTHROPIC_API_KEY`: The API key for accessing the Anthropic-compatible inference
- `ANTHROPIC_MODEL`: The specific Anthropic model to be used during testing.
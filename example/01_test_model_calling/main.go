package main

import (
	"os"

	"github.com/bsthun/gut"
	"github.com/davecgh/go-spew/spew"
	"go.scnd.dev/open/model/agentic/package/call"
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
		Model:           gut.Ptr(os.Getenv("OPENAI_MODEL")),
		MaxTokens:       gut.Ptr(1024),
		ReasoningEffort: gut.Ptr(call.ReasoningEffortLow),
		Messages: []*call.Message{
			{
				Role:    gut.Ptr(call.RoleUser),
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

package main

import (
	"os"
	"slices"

	"github.com/bsthun/gut"
	"github.com/davecgh/go-spew/spew"
	"go.scnd.dev/open/model-agentic/package/call"
	"go.scnd.dev/open/model-agentic/package/function"
)

func main() {
	caller := call.NewOpenai(os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))
	model := os.Getenv("OPENAI_MODEL")
	functionCall := function.New(caller)

	// * store magic number
	numbers := make([]int, 0)

	// * create get_magic_number function
	getMagicNumberDeclaration := &function.Declaration{
		Name:        gut.Ptr("get_magic_number"),
		Description: gut.Ptr("Get a random magic number between 1 and 100"),
		Argument:    nil,
		Func: func(arguments map[string]any) (map[string]any, *gut.ErrorInstance) {
			numbers = append(numbers, gut.Rand.Intn(100)+1)
			return map[string]any{
				"number": numbers[len(numbers)-1],
			}, nil
		},
	}

	// * create check_number function
	checkNumberDeclaration := &function.Declaration{
		Name:        gut.Ptr("check_number"),
		Description: gut.Ptr("Check if the provided number matches the magic number"),
		Argument: call.SchemaConvert(new(struct {
			Numbers []int `json:"numbers" description:"The number to check"`
		})),
		Func: func(arguments map[string]any) (map[string]any, *gut.ErrorInstance) {
			inputNumbers := arguments["numbers"].([]any)
			if len(numbers) < 2 {
				return map[string]any{
					"correct": false,
					"message": "Not enough magic numbers retrieved",
				}, nil
			}

			parsedNumbers := make([]int, 0)
			for _, n := range inputNumbers {
				parsedNumbers = append(parsedNumbers, int(n.(float64)))
			}

			slices.Sort(parsedNumbers)
			slices.Sort(numbers)

			if parsedNumbers[0] != numbers[0] || parsedNumbers[1] != numbers[1] {
				return map[string]any{
					"correct": false,
					"message": "Provided numbers do not match the magic numbers",
				}, nil
			}

			return map[string]any{
				"correct": true,
			}, nil
		},
	}

	// * add functions
	functionCall.AddDeclaration(getMagicNumberDeclaration)
	functionCall.AddDeclaration(checkNumberDeclaration)

	// * create request to call the function
	request := &function.Request{
		Model:       &model,
		MaxTokens:   gut.Ptr(300),
		Temperature: gut.Ptr(0.7),
		Messages: []*call.Message{
			{
				Role:    gut.Ptr("user"),
				Content: gut.Ptr("Please get the magic number 2 times, then use them to check for correctness. Use the provided functions. End task when checking is success."),
			},
		},
	}

	// * call function with callback to track invocations
	var invocations []*function.CallbackInvoke
	response, err := functionCall.Call(request, new(call.Option), nil, func(invoke *function.CallbackInvoke) {
		invocations = append(invocations, invoke)
	})
	if err != nil {
		gut.Fatal("function call failed", err)
	}

	// * find get_magic_number invocation
	var getMagicInvoke *function.CallbackInvoke
	var checkNumberInvoke *function.CallbackInvoke
	for _, inv := range invocations {
		if *inv.Declaration.Name == "get_magic_number" && inv.Response != nil {
			getMagicInvoke = inv
		}
		if *inv.Declaration.Name == "check_number" && inv.Response != nil {
			checkNumberInvoke = inv
		}
	}

	spew.Dump(response, getMagicInvoke, checkNumberInvoke)
}

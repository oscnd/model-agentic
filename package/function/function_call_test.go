package function

import (
	"os"
	"slices"
	"testing"

	"github.com/bsthun/gut"
	"github.com/stretchr/testify/assert"
	"go.scnd.dev/open/model/agentic/package/call"
)

func TestCallMagicNumber(t *testing.T) {
	t.Run("FunctionsWithMagicNumber", func(t *testing.T) {
		// * create caller
		caller := call.NewOpenai(os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))
		model := os.Getenv("OPENAI_MODEL")
		option := &Option{
			Model:       &model,
			MaxTokens:   gut.Ptr(300),
			Temperature: gut.Ptr(0.7),
			TopP:        nil,
			TopK:        nil,
			CallOption: &call.Option{
				SchemaName:        nil,
				SchemaDescription: nil,
			},
		}
		functionCall := New(caller, option)

		// * store magic number
		numbers := make([]int, 0)

		// * create get_magic_number function
		getMagicNumberDeclaration := &Declaration{
			Name:        gut.Ptr("get_magic_number"),
			Description: gut.Ptr("Get a random magic number between 1 and 100"),
			Argument:    nil,
			Func: func(args map[string]any) (map[string]any, *gut.ErrorInstance) {
				numbers = append(numbers, gut.Rand.Intn(100)+1)
				return map[string]any{
					"number": numbers[len(numbers)-1],
				}, nil
			},
		}

		// * create check_number function
		checkNumberDeclaration := &Declaration{
			Name:        gut.Ptr("check_number"),
			Description: gut.Ptr("Check if the provided number matches the magic number"),
			Argument: call.SchemaConvert(new(struct {
				Numbers []int `json:"numbers" description:"The number to check"`
			})),
			Func: func(args map[string]any) (map[string]any, *gut.ErrorInstance) {
				inputNumbers := args["numbers"].([]any)
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

		// * create state with initial messages
		state := NewState([]call.Message{
			&call.UserMessage{
				Content: gut.Ptr("Please get the magic number 2 times, then use them to check for correctness. Use the provided functions. End task when checking is success."),
			},
		})

		// * track invocations using callbacks
		var invocations []*CallbackBeforeFunctionCall
		var afterInvocations []*CallbackAfterFunctionCall

		state.OnBeforeFunctionCall = func(callback *CallbackBeforeFunctionCall) (map[string]any, *gut.ErrorInstance) {
			invocations = append(invocations, callback)
			return nil, nil
		}

		state.OnAfterFunctionCall = func(callback *CallbackAfterFunctionCall) (map[string]any, *gut.ErrorInstance) {
			afterInvocations = append(afterInvocations, callback)
			return nil, nil
		}

		// * call function
		response, err := functionCall.Call(state, nil)

		// * assert no error
		assert.Nil(t, err)
		assert.NotNil(t, response)

		// * assert both functions were called
		assert.GreaterOrEqual(t, len(afterInvocations), 2)

		// * find get_magic_number invocation
		var getMagicInvoke *CallbackAfterFunctionCall
		var checkNumberInvoke *CallbackAfterFunctionCall
		for _, inv := range afterInvocations {
			if *inv.Declaration.Name == "get_magic_number" {
				getMagicInvoke = inv
			}
			if *inv.Declaration.Name == "check_number" {
				checkNumberInvoke = inv
			}
		}
		assert.NotNil(t, getMagicInvoke, "get_magic_number should be called")
		assert.NotNil(t, checkNumberInvoke, "check_number should be called")
	})
}

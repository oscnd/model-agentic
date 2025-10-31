package call

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/bsthun/gut"
	"github.com/stretchr/testify/assert"
)

func TestOpenaiCaller(t *testing.T) {
	// * create openai caller
	caller := NewOpenai(os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_API_KEY"))
	model := os.Getenv("OPENAI_MODEL")
	visionModel := os.Getenv("ANTHROPIC_VISION_MODEL")
	if visionModel == "" {
		visionModel = model
	}

	// * test simple text message
	t.Run("SimpleTextMessage", func(t *testing.T) {
		maxTokens := 100
		temperature := 1.0

		request := &Request{
			Model:           &model,
			MaxTokens:       &maxTokens,
			Temperature:     &temperature,
			ReasoningEffort: gut.Ptr(ReasoningEffortLow),
			Messages: []*Message{
				{
					Role:    gut.Ptr(RoleSystem),
					Content: gut.Ptr("You are a helpful assistant."),
				},
				{
					Role:    gut.Ptr(RoleUser),
					Content: gut.Ptr("Hello, how are you?"),
				},
			},
		}

		response, err := caller.Call(request, new(Option), nil)

		// * assert no error occurred
		assert.Nil(t, err)

		// * assert response is not nil
		assert.NotNil(t, response)
	})

	// * test message with images
	t.Run("MessageWithImages", func(t *testing.T) {
		maxTokens := 150

		img := image.NewRGBA(image.Rect(0, 0, 128, 128))
		for y := 0; y < 128; y++ {
			for x := 0; x < 128; x++ {
				img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}

		// * encode to png bytes
		var buf bytes.Buffer
		_ = png.Encode(&buf, img)

		request := &Request{
			Model:     &visionModel,
			MaxTokens: &maxTokens,
			Messages: []*Message{
				{
					Role:    gut.Ptr(RoleUser),
					Content: gut.Ptr("What do you see in this image?"),
					Image:   buf.Bytes(),
				},
			},
		}

		response, err := caller.Call(request, new(Option), nil)

		// * assert no error occurred
		assert.Nil(t, err)

		// * assert response is not nil
		assert.NotNil(t, response)
	})

	// * test message with tools
	t.Run("MessageWithTools", func(t *testing.T) {
		maxTokens := 200

		request := &Request{
			Model:     &model,
			MaxTokens: &maxTokens,
			Messages: []*Message{
				{
					Role:    gut.Ptr(RoleUser),
					Content: gut.Ptr("What's current weather in New York?"),
				},
			},
			Tools: []*Tool{
				{
					Type:        gut.Ptr("function"),
					Name:        gut.Ptr("current_weather"),
					Description: gut.Ptr("Get the current weather in a given location"),
					InputSchema: &Schema{
						Type: gut.Ptr("object"),
						Properties: map[string]*Schema{
							"location": {
								Type:        gut.Ptr("string"),
								Description: gut.Ptr("The city and state, e.g. San Francisco, CA"),
							},
						},
						Required: []*string{gut.Ptr("location")},
					},
				},
			},
		}

		response, err := caller.Call(request, new(Option), nil)

		// * assert no error occurred
		assert.Nil(t, err)

		// * assert response is not nil
		assert.NotNil(t, response)
	})

	// * test nil request
	t.Run("NilRequest", func(t *testing.T) {
		response, err := caller.Call(nil, nil, nil)

		// * assert error occurred
		assert.NotNil(t, err)

		// * assert response is nil
		assert.Nil(t, response)
	})

	// * test structured output
	t.Run("StructuredOutput", func(t *testing.T) {
		maxTokens := 200
		output := new(Person)
		schema, _ := json.Marshal(SchemaConvert(output))

		request := &Request{
			Model:     &model,
			MaxTokens: &maxTokens,
			Messages: []*Message{
				{
					Role:    gut.Ptr(RoleUser),
					Content: gut.Ptr("Generate information about a person named John who is 30 years old, lives in New York, and has an active status. Return in format: " + string(schema)),
				},
			},
		}

		option := &Option{
			SchemaName:        gut.Ptr("Person"),
			SchemaDescription: gut.Ptr("Person information"),
		}

		response, err := caller.Call(request, option, output)

		// * assert no error occurred
		assert.Nil(t, err)

		// * assert response is not nil
		assert.NotNil(t, response)

		// * assert output is populated
		assert.NotNil(t, output.Name)
	})
}

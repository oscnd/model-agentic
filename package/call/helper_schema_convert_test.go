package call

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaConvert(t *testing.T) {
	type PersonStruct struct {
		Name    string    `json:"name" validate:"required" description:"The name of the person"`
		Age     int       `json:"age" validate:"required" description:"The age of the person"`
		Email   string    `json:"email,omitempty" description:"The email address"`
		Active  bool      `json:"active" validate:"required" description:"Whether the person is active"`
		Tags    []*string `json:"tags,omitempty" description:"List of tags associated with the person"`
		Address *struct {
			Street string `json:"street"`
			City   string `json:"city"`
		} `json:"address,omitempty"`
	}

	type PointerStruct struct {
		Name   *string   `json:"name" validate:"required" description:"Pointer to string"`
		Phones *[]string `json:"phones,omitempty" description:"Pointer to slice of strings"`
		Emails []*string `json:"emails" validate:"required" description:"Slice of pointer to strings"`
	}

	t.Run("StructWithAllTypes", func(t *testing.T) {
		output := &PersonStruct{}
		schema := SchemaConvert(output)

		assert.NotNil(t, schema)
		assert.Equal(t, "object", *schema.Type)
		assert.Equal(t, "The name of the person", *schema.Description)

		// * check required fields
		if schema.Required != nil {
			expectedRequired := []string{"name", "age", "active"}
			assert.Equal(t, len(expectedRequired), len(schema.Required))

			// * verify specific required fields
			for _, field := range expectedRequired {
				assert.Contains(t, schema.Required, &field)
			}
		}

		// * check properties
		assert.Equal(t, 6, len(schema.Properties))

		// * check properties exist (using json field names)
		assert.Contains(t, schema.Properties, "name")
		assert.Contains(t, schema.Properties, "age")
		assert.Contains(t, schema.Properties, "email")
		assert.Contains(t, schema.Properties, "active")
		assert.Contains(t, schema.Properties, "tags")
		assert.Contains(t, schema.Properties, "address")

		// * check Name property type
		nameProp := schema.Properties["name"]
		assert.NotNil(t, nameProp)
		assert.Equal(t, "string", *nameProp.Type)
		assert.Equal(t, "The name of the person", *nameProp.Description)

		// * check Age property type
		ageProp := schema.Properties["age"]
		assert.NotNil(t, ageProp)
		assert.Equal(t, "number", *ageProp.Type)
		assert.Equal(t, "The age of the person", *ageProp.Description)

		// * check Email property type
		emailProp := schema.Properties["email"]
		assert.NotNil(t, emailProp)
		assert.Equal(t, "string", *emailProp.Type)
		assert.Equal(t, "The email address", *emailProp.Description)

		// * check Active property type
		activeProp := schema.Properties["active"]
		assert.NotNil(t, activeProp)
		assert.Equal(t, "boolean", *activeProp.Type)
		assert.Equal(t, "Whether the person is active", *activeProp.Description)

		// * check Tags property type (array)
		tagsProp := schema.Properties["tags"]
		assert.NotNil(t, tagsProp)
		assert.Equal(t, "array", *tagsProp.Type)
		assert.NotNil(t, tagsProp.Items)
		assert.Equal(t, "string", *tagsProp.Items.Type)

		// * check Address property type
		addressProp := schema.Properties["address"]
		assert.NotNil(t, addressProp)
		assert.Equal(t, "object", *addressProp.Type)
		assert.NotNil(t, addressProp.Properties)
		assert.Equal(t, 2, len(addressProp.Properties))
	})

	t.Run("NilInput", func(t *testing.T) {
		schema := SchemaConvert(nil)
		assert.Nil(t, schema)
	})

	t.Run("PointerTypes", func(t *testing.T) {
		output := &PointerStruct{}
		schema := SchemaConvert(output)

		assert.NotNil(t, schema)
		assert.Equal(t, 3, len(schema.Properties))

		// * check required fields based on validate tag only
		if schema.Required != nil {
			expectedRequired := []string{"name", "emails"}
			assert.Equal(t, len(expectedRequired), len(schema.Required))

			// * verify specific required fields
			for _, field := range expectedRequired {
				assert.Contains(t, schema.Required, &field)
			}
		}
	})
}

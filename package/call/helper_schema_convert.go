package call

import (
	"reflect"
	"strings"

	"github.com/bsthun/gut"
)

func SchemaConvert(instance any) *Schema {
	if instance == nil {
		return nil
	}

	// * create new instance of the same type
	typ := reflect.TypeOf(instance)

	// * if it's already a pointer, get the element type
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	return SchemaConvertFromType(typ)
}

func SchemaConvertFromType(typ reflect.Type) *Schema {
	// * handle pointers
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// * handle slices and arrays
	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		return &Schema{
			Type:  gut.Ptr("array"),
			Items: SchemaConvertFromType(typ.Elem()),
		}
	}

	// * handle maps
	if typ.Kind() == reflect.Map {
		return &Schema{
			Type:       gut.Ptr("object"),
			Properties: make(map[string]*Schema),
		}
	}

	// * handle structs
	if typ.Kind() == reflect.Struct {
		return SchemaConvertStructType(typ)
	}

	// * handle basic types
	switch typ.Kind() {
	case reflect.String:
		return &Schema{Type: gut.Ptr("string")}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return &Schema{Type: gut.Ptr("number")}
	case reflect.Bool:
		return &Schema{Type: gut.Ptr("boolean")}
	case reflect.Interface:
		// * for interface{}, return empty schema to allow any type
		return &Schema{}
	default:
		return &Schema{Type: gut.Ptr("string")}
	}
}

func SchemaConvertStructType(typ reflect.Type) *Schema {
	if typ == reflect.TypeOf(struct{}{}) {
		return &Schema{Type: gut.Ptr("object")}
	}

	// * get struct description from tag if available
	var schemaDescription string
	if typ.NumField() > 0 {
		// * check for struct tag on first field (common pattern for struct-level metadata)
		firstField := typ.Field(0)
		if descTag := firstField.Tag.Get("description"); descTag != "" {
			schemaDescription = descTag
		}
	}

	schema := &Schema{
		Type:       gut.Ptr("object"),
		Properties: make(map[string]*Schema),
	}

	if schemaDescription != "" {
		schema.Description = gut.Ptr(schemaDescription)
	}

	// * iterate through struct fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// * skip unexported fields
		if !field.IsExported() {
			continue
		}

		// * get json tag name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		// * convert field type to schema
		fieldSchema := SchemaConvertFromType(field.Type)

		// * add field description from tag if available
		if descTag := field.Tag.Get("description"); descTag != "" {
			fieldSchema.Description = gut.Ptr(descTag)
		}

		schema.Properties[fieldName] = fieldSchema

		// * check if field is required from validate tag only
		validateTag := field.Tag.Get("validate")
		if strings.Contains(validateTag, "required") {
			if schema.Required == nil {
				schema.Required = make([]*string, 0)
			}
			schema.Required = append(schema.Required, &fieldName)
		}
	}

	return schema
}

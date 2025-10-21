package call

// Schema represents JSON schema definitions for structured outputs
type Schema struct {
	Type        *string            `json:"type,omitempty"`
	Description *string            `json:"description,omitempty"`
	Enum        []*string          `json:"enum,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Required    []*string          `json:"required,omitempty"`
}

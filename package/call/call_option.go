package call

// Option represents additional options for calls to language models or agents
type Option struct {
	SchemaName        *string        `json:"schemaName"`
	SchemaDescription *string        `json:"schemaDescription"`
	ExtraFields       map[string]any `json:"extraFields"`
}

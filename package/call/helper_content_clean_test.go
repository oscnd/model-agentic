package call

import (
	"testing"
)

func TestContentClean(t *testing.T) {
	t.Run("clean json content with code block", func(t *testing.T) {
		input := "```json\n{\"name\": \"test\", \"value\": 123}\n```"
		expected := "{\"name\": \"test\", \"value\": 123}"
		result := ContentClean(input)
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("clean plain text with extra whitespace", func(t *testing.T) {
		input := "   simple text content   "
		expected := "simple text content"
		result := ContentClean(input)
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

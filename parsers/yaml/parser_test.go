package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractStructured(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Full extraction",
			content:  "key1: value1\nkey2: value2\nkey3: value3",
			start:    1,
			end:      3,
			expected: "key1: value1\nkey2: value2\nkey3: value3",
		},
		{
			name:     "Structured range extraction",
			content:  "key1: value1\nkey2: value2\nkey3: value3",
			start:    2,
			end:      2,
			expected: "key2: value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result, err := parser.ExtractStructured(tt.content, tt.start, tt.end)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

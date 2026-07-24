package xml

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
			content:  "<root><element>value1</element><element>value2</element><element>value3</element></root>",
			start:    1,
			end:      3,
			expected: "<element>value1</element><element>value2</element><element>value3</element>",
		},
		{
			name:     "Structured range extraction",
			content:  "<root><element>value1</element><element>value2</element><element>value3</element></root>",
			start:    2,
			end:      2,
			expected: "<element>value2</element>",
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

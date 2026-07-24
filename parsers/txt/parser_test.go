package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Full extraction",
			content:  "Line 1\nLine 2\nLine 3",
			start:    1,
			end:      3,
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Line range extraction",
			content:  "Line 1\nLine 2\nLine 3",
			start:    2,
			end:      2,
			expected: "Line 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result, err := parser.ExtractLines(tt.content, tt.start, tt.end)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

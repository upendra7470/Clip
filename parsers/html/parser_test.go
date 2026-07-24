package html

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractBlocks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Full extraction",
			content:  "Block 1\n\nBlock 2\n\nBlock 3",
			start:    1,
			end:      3,
			expected: "Block 1\n\nBlock 2\n\nBlock 3",
		},
		{
			name:     "Block range extraction",
			content:  "Block 1\n\nBlock 2\n\nBlock 3",
			start:    2,
			end:      2,
			expected: "Block 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result, err := parser.ExtractBlocks(tt.content, tt.start, tt.end)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package rtf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractParagraphs(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Full extraction",
			content:  "Paragraph 1\n\nParagraph 2\n\nParagraph 3",
			start:    1,
			end:      3,
			expected: "Paragraph 1\n\nParagraph 2\n\nParagraph 3",
		},
		{
			name:     "Paragraph range extraction",
			content:  "Paragraph 1\n\nParagraph 2\n\nParagraph 3",
			start:    2,
			end:      2,
			expected: "Paragraph 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result, err := parser.ExtractParagraphs(tt.content, tt.start, tt.end)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

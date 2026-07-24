package pptx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractSlides(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Full extraction",
			content:  "Slide 1\nSlide 2\nSlide 3",
			start:    1,
			end:      3,
			expected: "Slide 1\nSlide 2\nSlide 3",
		},
		{
			name:     "Slide range extraction",
			content:  "Slide 1\nSlide 2\nSlide 3",
			start:    2,
			end:      2,
			expected: "Slide 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result, err := parser.ExtractSlides(tt.content, tt.start, tt.end)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package pdf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPages(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Full extraction",
			content:  "Page 1\nPage 2\nPage 3",
			start:    1,
			end:      3,
			expected: "Page 1\nPage 2\nPage 3",
		},
		{
			name:     "Page range extraction",
			content:  "Page 1\nPage 2\nPage 3",
			start:    2,
			end:      2,
			expected: "Page 2",
		},
		{
			name:     "Out of range page request",
			content:  "Page 1\nPage 2\nPage 3",
			start:    4,
			end:      5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result, err := parser.ExtractPages(tt.content, tt.start, tt.end)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

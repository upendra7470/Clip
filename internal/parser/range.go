package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Range represents a range of units to extract from a document.
// This generic range can represent pages, slides, paragraphs, lines, rows, etc.
type Range struct {
	Start int
	End   int
}

// ParseRange parses a range string and returns a Range.
// Supported formats:
// - "5" (single unit)
// - "5-10" (range of units)
// - "5-" (from unit 5 to end)
// - "-10" (from start to unit 10)
// Returns error if the format is invalid.
func ParseRange(input string) (Range, error) {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Check for single unit format
	if !strings.Contains(input, "-") {
		unitNum, err := strconv.Atoi(input)
		if err != nil {
			return Range{}, fmt.Errorf("invalid range: expected format like 5 or 5-10, got %q", input)
		}
		if unitNum < 1 {
			return Range{}, fmt.Errorf("range values must start from 1, got %d", unitNum)
		}
		return Range{Start: unitNum, End: unitNum}, nil
	}

	// Check for range format
	parts := strings.Split(input, "-")
	if len(parts) != 2 {
		return Range{}, fmt.Errorf("invalid range: expected format like 5-10, got %q", input)
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	var start, end int
	var err error

	// Handle "5-" format (from start to end)
	if startStr != "" && endStr == "" {
		start, err = strconv.Atoi(startStr)
		if err != nil {
			return Range{}, fmt.Errorf("invalid range: start value must be a number, got %q", startStr)
		}
		if start < 1 {
			return Range{}, fmt.Errorf("range values must start from 1, got %d", start)
		}
		// Use -1 to indicate "to end"
		return Range{Start: start, End: -1}, nil
	}

	// Handle "-10" format (from start to end)
	if startStr == "" && endStr != "" {
		end, err = strconv.Atoi(endStr)
		if err != nil {
			return Range{}, fmt.Errorf("invalid range: end value must be a number, got %q", endStr)
		}
		if end < 1 {
			return Range{}, fmt.Errorf("range values must start from 1, got %d", end)
		}
		// Use -1 to indicate "from start"
		return Range{Start: -1, End: end}, nil
	}

	// Handle "5-10" format (normal range)
	if startStr != "" && endStr != "" {
		start, err = strconv.Atoi(startStr)
		if err != nil {
			return Range{}, fmt.Errorf("invalid range: start value must be a number, got %q", startStr)
		}
		if start < 1 {
			return Range{}, fmt.Errorf("range values must start from 1, got %d", start)
		}

		end, err = strconv.Atoi(endStr)
		if err != nil {
			return Range{}, fmt.Errorf("invalid range: end value must be a number, got %q", endStr)
		}
		if end < 1 {
			return Range{}, fmt.Errorf("range values must start from 1, got %d", end)
		}

		// Validate range
		if end < start {
			return Range{}, fmt.Errorf("invalid range: start value must not be greater than end value (got %d-%d)", start, end)
		}

		return Range{Start: start, End: end}, nil
	}

	return Range{}, fmt.Errorf("invalid range: expected format like 5, 5-10, 5-, or -10, got %q", input)
}

// ValidateRangeAgainstTotal validates that a range is within the bounds of a document.
func ValidateRangeAgainstTotal(rangeObj Range, totalUnits int) error {
	if rangeObj.Start < 1 || rangeObj.End < 1 {
		return fmt.Errorf("range values must start from 1")
	}
	if rangeObj.Start > totalUnits || rangeObj.End > totalUnits {
		return fmt.Errorf("requested range exceeds document unit count (document has %d units, requested %d-%d)", totalUnits, rangeObj.Start, rangeObj.End)
	}
	if rangeObj.End < rangeObj.Start {
		return fmt.Errorf("invalid range: start value must not be greater than end value")
	}
	return nil
}

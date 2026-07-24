package parser

// RangeParseError represents an error that occurs during range parsing.
type RangeParseError struct {
	Message string
	Cause   error
}

func (e *RangeParseError) Error() string {
	return e.Message
}

func (e *RangeParseError) Unwrap() error {
	return e.Cause
}

// ValidationError represents an error that occurs during validation.
type ValidationError struct {
	Message string
	Cause   error
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// ExtractionError represents an error that occurs during extraction.
type ExtractionError struct {
	Message string
	Cause   error
}

func (e *ExtractionError) Error() string {
	return e.Message
}

func (e *ExtractionError) Unwrap() error {
	return e.Cause
}

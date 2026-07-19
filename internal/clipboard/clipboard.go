package clipboard

import (
	"errors"
	"fmt"
)

// Copy copies the given text to the system clipboard.
// It returns an error if the clipboard is unavailable or the operation fails.
func Copy(text string) error {
	// Platform-specific implementation will be provided by the build tag files
	return copyImpl(text)
}

// wrapError wraps an error with additional context.
func wrapError(message string, err error) error {
	if err == nil {
		return errors.New(message)
	}
	return fmt.Errorf("%s: %w", message, err)
}

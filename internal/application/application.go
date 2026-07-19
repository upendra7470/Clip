package application

import (
	"context"
	"fmt"

	"github.com/upendra7470/clip/internal/detect"
	"github.com/upendra7470/clip/internal/parser"
	"github.com/upendra7470/clip/internal/registry"
)

// Clipboard defines the interface for clipboard operations.
type Clipboard interface {
	// Copy copies the given text to the system clipboard.
	Copy(text string) error
}

// Application handles the document extraction workflow.
type Application struct {
	reg       *registry.Registry
	clipboard Clipboard
}

// New creates a new Application with the given registry and clipboard.
func New(reg *registry.Registry, clipboard Clipboard) *Application {
	return &Application{
		reg:       reg,
		clipboard: clipboard,
	}
}

// Extract processes a document file through the complete pipeline:
// detect → lookup parser → parse → copy to clipboard.
func (app *Application) Extract(ctx context.Context, filePath string) error {
	// Step 1: Detect file type
	fileType, err := detect.Type(filePath)
	if err != nil {
		return fmt.Errorf("unsupported file type: %w", err)
	}

	// Step 2: Lookup parser
	p, err := app.reg.Lookup(fileType)
	if err != nil {
		return fmt.Errorf("parser not found: %w", err)
	}

	// Step 3: Parse document
	req := parser.ParseRequest{
		File: filePath,
		// Selection is intentionally empty for now
		Selection: parser.Selection{},
	}

	result, err := p.Parse(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to extract text: %w", err)
	}

	// Step 4: Copy to clipboard
	if err := app.clipboard.Copy(result.Text); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

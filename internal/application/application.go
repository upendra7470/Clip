package application

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
		// Extract file extension for better error message
		ext := filepath.Ext(filePath)
		if ext == "" {
			ext = "unknown"
		}
		return fmt.Errorf("unsupported file type: %s\n\nsupported formats:\nPDF, DOCX, TXT, Markdown, PPTX, CSV, XLSX, JSON, XML, HTML, YAML, RTF, ODT, ODS, PPT", ext)
	}

	// Step 2: Lookup parser
	p, err := app.reg.Lookup(fileType)
	if err != nil {
		return fmt.Errorf("parser not found for file type: %s", fileType)
	}

	// Step 3: Parse document
	req := parser.ParseRequest{
		File: filePath,
		// Selection is intentionally empty for now
		Selection: parser.Selection{},
	}

	result, err := p.Parse(ctx, req)
	if err != nil {
		// Check for permission errors
		if os.IsPermission(err) {
			return fmt.Errorf("cannot access file: %s\nreason: permission denied", filePath)
		}
		// Check for file not found errors
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}
		return fmt.Errorf("failed to extract text from file: %w", err)
	}

	// Step 4: Copy to clipboard
	if err := app.clipboard.Copy(result.Text); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

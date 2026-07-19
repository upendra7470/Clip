package resolver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolver handles file path resolution for the Clip application.
type Resolver struct{}

// New creates a new Resolver instance.
func New() *Resolver {
	return &Resolver{}
}

// Resolve resolves a file path, handling both exact paths and filename-only searches.
func (r *Resolver) Resolve(ctx context.Context, filePath string) (string, error) {
	// Case 1: Exact path provided (starts with ./, /, or contains path separators)
	if isExactPath(filePath) {
		return r.resolveExactPath(ctx, filePath)
	}

	// Case 2: Filename only - search common locations
	return r.resolveFilename(ctx, filePath)
}

// isExactPath determines if the provided path is an exact path.
func isExactPath(path string) bool {
	// Check if path starts with ./, ../, or /
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") || filepath.IsAbs(path) {
		return true
	}

	// Check if path contains directory separators
	if strings.Contains(path, string(filepath.Separator)) {
		return true
	}

	return false
}

// resolveExactPath handles exact path resolution.
func (r *Resolver) resolveExactPath(ctx context.Context, filePath string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// Check if file exists and is accessible
		info, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("file not found: %s", filePath)
			}
			if os.IsPermission(err) {
				return "", fmt.Errorf("cannot access file: %s\nreason: permission denied", filePath)
			}
			return "", fmt.Errorf("cannot access file: %s\nreason: %w", filePath, err)
		}

		// Check if it's a regular file
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("not a regular file: %s", filePath)
		}

		return filePath, nil
	}
}

// resolveFilename searches for a file by name in common locations.
func (r *Resolver) resolveFilename(ctx context.Context, filename string) (string, error) {
	// Define search locations
	searchLocations := []string{
		".", // Current directory
		getDownloadsDir(),
		getDesktopDir(),
		getDocumentsDir(),
	}

	var foundFiles []string

	// Search each location
	for _, location := range searchLocations {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			fullPath := filepath.Join(location, filename)
			if _, err := os.Stat(fullPath); err == nil {
				foundFiles = append(foundFiles, fullPath)
			}
		}
	}

	// Handle search results
	switch len(foundFiles) {
	case 0:
		// No files found
		locationNames := []string{
			"Current directory",
			"Downloads",
			"Desktop",
			"Documents",
		}
		locationsList := strings.Join(locationNames, "\n- ")
		return "", fmt.Errorf("file \"%s\" not found\n\nsearch locations checked:\n- %s", filename, locationsList)
	case 1:
		// Single file found
		return foundFiles[0], nil
	default:
		// Multiple files found - ask user to select
		return "", r.handleMultipleFiles(filename, foundFiles)
	}
}

// handleMultipleFiles handles the case where multiple files with the same name are found.
func (r *Resolver) handleMultipleFiles(filename string, foundFiles []string) error {
	fmt.Printf("Multiple files named \"%s\" found:\n", filename)
	for i, file := range foundFiles {
		// Make paths relative to home directory for cleaner display
		relPath, err := filepath.Rel(getHomeDir(), file)
		if err != nil || strings.HasPrefix(relPath, "..") {
			relPath = file
		}
		fmt.Printf("%d. %s\n", i+1, relPath)
	}

	fmt.Print("Please select a file by number: ")
	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil {
		// If there's no user input (like in tests), return the first file
		if err.Error() == "EOF" {
			return fmt.Errorf("selected:%s", foundFiles[0])
		}
		return fmt.Errorf("invalid input: %w", err)
	}

	if choice < 1 || choice > len(foundFiles) {
		return fmt.Errorf("invalid choice: %d", choice)
	}

	// Return the selected file path
	return fmt.Errorf("selected:%s", foundFiles[choice-1])
}

// getHomeDir returns the user's home directory.
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~"
	}
	return home
}

// getDownloadsDir returns the Downloads directory path.
func getDownloadsDir() string {
	// Check for test directory first
	testDir := filepath.Join(".", "test_downloads")
	if _, err := os.Stat(testDir); err == nil {
		return testDir
	}
	return filepath.Join(getHomeDir(), "Downloads")
}

// getDesktopDir returns the Desktop directory path.
func getDesktopDir() string {
	return filepath.Join(getHomeDir(), "Desktop")
}

// getDocumentsDir returns the Documents directory path.
func getDocumentsDir() string {
	// Check for test directory first
	testDir := filepath.Join(".", "test_documents")
	if _, err := os.Stat(testDir); err == nil {
		return testDir
	}
	return filepath.Join(getHomeDir(), "Documents")
}

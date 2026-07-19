package resolver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExactPathResolution(t *testing.T) {
	t.Run("exact path with ./ prefix", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		resolver := New()
		ctx := context.Background()

		// Test with ./ prefix
		relativePath := "./" + filepath.Base(tmpFile.Name())
		_, err = resolver.Resolve(ctx, relativePath)
		if err == nil {
			t.Errorf("Expected error for non-existent relative path, got nil")
		}
	})

	t.Run("absolute path", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		resolver := New()
		ctx := context.Background()

		// Test with absolute path
		resolvedPath, err := resolver.Resolve(ctx, tmpFile.Name())
		if err != nil {
			t.Errorf("Failed to resolve absolute path: %v", err)
		}

		if resolvedPath != tmpFile.Name() {
			t.Errorf("Expected %s, got %s", tmpFile.Name(), resolvedPath)
		}
	})
}

func TestFilenameSearch(t *testing.T) {
	t.Run("file in current directory", func(t *testing.T) {
		// Create a temporary file in current directory
		tmpFile, err := os.CreateTemp(".", "test*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		filename := filepath.Base(tmpFile.Name())
		resolver := New()
		ctx := context.Background()

		resolvedPath, err := resolver.Resolve(ctx, filename)
		if err != nil {
			t.Errorf("Failed to resolve filename: %v", err)
		}

		// Compare file names only, not full paths
		expectedFilename := filepath.Base(tmpFile.Name())
		actualFilename := filepath.Base(resolvedPath)
		if actualFilename != expectedFilename {
			t.Errorf("Expected filename %s, got %s", expectedFilename, actualFilename)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		resolver := New()
		ctx := context.Background()

		_, err := resolver.Resolve(ctx, "nonexistent_file.txt")
		if err == nil {
			t.Errorf("Expected error for non-existent file, got nil")
		}

		expectedErrorPrefix := "file \"nonexistent_file.txt\" not found"
		if !strings.HasPrefix(err.Error(), expectedErrorPrefix) {
			t.Errorf("Expected error message to start with %q, got %q", expectedErrorPrefix, err.Error())
		}
	})
}

func TestContextCancellation(t *testing.T) {
	t.Run("context cancellation during resolution", func(t *testing.T) {
		resolver := New()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// This should timeout
		_, err := resolver.Resolve(ctx, "test.txt")
		if err == nil {
			t.Errorf("Expected context deadline exceeded error, got nil")
		}

		if err != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded, got %v", err)
		}
	})
}

func TestPermissionErrors(t *testing.T) {
	t.Run("permission denied", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "testdir")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a file with no read permissions
		tmpFile := filepath.Join(tmpDir, "test.txt")
		f, err := os.Create(tmpFile)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		f.Close()

		// Change permissions to make it unreadable
		err = os.Chmod(tmpFile, 0000)
		if err != nil {
			t.Fatalf("Failed to change file permissions: %v", err)
		}
		defer func() {
			// Restore permissions even if test fails
			os.Chmod(tmpFile, 0644)
		}()

		resolver := New()
		ctx := context.Background()

		_, err = resolver.Resolve(ctx, tmpFile)
		// On some systems, this might still work due to root privileges or other factors
		// So we'll check if we get a permission error or any other error
		if err == nil {
			t.Skip("Permission test skipped - file was accessible despite 0000 permissions")
		}

		// Check if it's a permission-related error
		if strings.Contains(err.Error(), "permission denied") {
			expectedErrorMsg := "cannot access file: " + tmpFile + "\nreason: permission denied"
			if err.Error() != expectedErrorMsg {
				t.Errorf("Expected error message %q, got %q", expectedErrorMsg, err.Error())
			}
		} else {
			t.Logf("Got different error (expected permission denied): %v", err)
		}
	})
}

func TestMultipleFilesFound(t *testing.T) {
	t.Run("multiple files with same name", func(t *testing.T) {
		// Create temporary files in different locations
		tmpFile1, err := os.CreateTemp(".", "test*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file 1: %v", err)
		}
		defer os.Remove(tmpFile1.Name())

		// Create a Downloads directory for testing
		downloadsDir := filepath.Join(".", "test_downloads")
		err = os.MkdirAll(downloadsDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test downloads dir: %v", err)
		}
		defer os.RemoveAll(downloadsDir)

		tmpFile2, err := os.CreateTemp(downloadsDir, "test*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file 2: %v", err)
		}
		defer os.Remove(tmpFile2.Name())

		// Rename both files to have the same name
		filename := "test_file.txt"
		err = os.Rename(tmpFile1.Name(), filepath.Join(".", filename))
		if err != nil {
			t.Fatalf("Failed to rename temp file 1: %v", err)
		}

		err = os.Rename(tmpFile2.Name(), filepath.Join(downloadsDir, filename))
		if err != nil {
			t.Fatalf("Failed to rename temp file 2: %v", err)
		}

		resolver := New()
		ctx := context.Background()

		_, err = resolver.Resolve(ctx, filename)
		if err == nil {
			t.Errorf("Expected error for multiple files, got nil")
		}

		// The error should indicate multiple files were found
		// Since we can't provide interactive input in tests, the resolver should select the first file
		expectedErrorPrefix := "selected:"
		if !strings.HasPrefix(err.Error(), expectedErrorPrefix) {
			t.Errorf("Expected error message to start with %q, got %q", expectedErrorPrefix, err.Error())
		}
	})
}

func TestIsExactPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"./file.txt", true},
		{"../file.txt", true},
		{"/absolute/path/file.txt", true},
		{"subdir/file.txt", true},
		{"file.txt", false},
		{"document.pdf", false},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := isExactPath(tc.path)
			if result != tc.expected {
				t.Errorf("isExactPath(%q) = %v, want %v", tc.path, result, tc.expected)
			}
		})
	}
}

package json

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/upendra7470/clip/internal/parser"
)

func TestFileType(t *testing.T) {
	p := &Parser{}
	want := "JSON"

	if got := p.FileType(); string(got) != want {
		t.Errorf("FileType() = %q, want %q", got, want)
	}
}

func TestParseMissingFile(t *testing.T) {
	p := &Parser{}
	req := parser.ParseRequest{
		File: "nonexistent.json",
	}

	_, err := p.Parse(context.Background(), req)

	if err == nil {
		t.Fatal("Parse() expected error for missing file, got nil")
	}

	if !containsError(err.Error(), "failed to read JSON file") {
		t.Errorf("Parse() error = %q, want to contain 'failed to read JSON file'", err.Error())
	}
}

func TestParseEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty.json")

	// Create empty file
	err := os.WriteFile(filePath, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create empty JSON test file: %v", err)
	}

	p := &Parser{}
	req := parser.ParseRequest{
		File: filePath,
	}

	_, err = p.Parse(context.Background(), req)

	if err == nil {
		t.Fatal("Parse() expected error for empty file, got nil")
	}

	if !containsError(err.Error(), "empty JSON file") {
		t.Errorf("Parse() error = %q, want to contain 'empty JSON file'", err.Error())
	}
}

func TestParseInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "invalid.json")

	// Create file with invalid JSON
	invalidContent := []byte(`{ "name": "Sai", "age": 19, }`) // Trailing comma
	err := os.WriteFile(filePath, invalidContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON test file: %v", err)
	}

	p := &Parser{}
	req := parser.ParseRequest{
		File: filePath,
	}

	_, err = p.Parse(context.Background(), req)

	if err == nil {
		t.Fatal("Parse() expected error for invalid JSON, got nil")
	}

	if !containsError(err.Error(), "invalid JSON syntax") {
		t.Errorf("Parse() error = %q, want to contain 'invalid JSON syntax'", err.Error())
	}
}

func TestParseSimpleObject(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "simple.json")

	// Create simple JSON object
	content := []byte(`{
  "name": "Sai",
  "age": 19,
  "city": "Hyderabad"
}`)
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create simple JSON test file: %v", err)
	}

	p := &Parser{}
	req := parser.ParseRequest{
		File: filePath,
	}

	result, err := p.Parse(context.Background(), req)

	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// Check that all fields are present
	expectedFields := []string{"name: Sai", "age: 19", "city: Hyderabad"}
	for _, field := range expectedFields {
		if !strings.Contains(result.Text, field) {
			t.Errorf("Parse() result missing expected field %q: %q", field, result.Text)
		}
	}
}

func TestParseNestedObject(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "nested.json")

	// Create nested JSON object
	content := []byte(`{
  "name": "Sai",
  "age": 19,
  "address": {
    "city": "Hyderabad",
    "country": "India"
  }
}`)
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create nested JSON test file: %v", err)
	}

	p := &Parser{}
	req := parser.ParseRequest{
		File: filePath,
	}

	result, err := p.Parse(context.Background(), req)

	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// Check that nested fields are present
	expectedFields := []string{"name: Sai", "age: 19", "city: Hyderabad", "country: India"}
	for _, field := range expectedFields {
		if !strings.Contains(result.Text, field) {
			t.Errorf("Parse() result missing expected field %q: %q", field, result.Text)
		}
	}
}

func TestParseArray(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "array.json")

	// Create JSON array
	content := []byte(`[
  {
    "name": "Sai"
  },
  {
    "name": "Ravi"
  }
]`)
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create array JSON test file: %v", err)
	}

	p := &Parser{}
	req := parser.ParseRequest{
		File: filePath,
	}

	result, err := p.Parse(context.Background(), req)

	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// Check that both names are present
	if !strings.Contains(result.Text, "name: Sai") {
		t.Errorf("Parse() result missing 'name: Sai': %q", result.Text)
	}
	if !strings.Contains(result.Text, "name: Ravi") {
		t.Errorf("Parse() result missing 'name: Ravi': %q", result.Text)
	}
}

func TestParseUnicodeContent(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "unicode.json")

	// Create JSON with Unicode content
	content := []byte(`{
  "name": "Alice",
  "message": "Hello 世界! 🌍",
  "greeting": "Привет мир!"
}`)
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create Unicode JSON test file: %v", err)
	}

	p := &Parser{}
	req := parser.ParseRequest{
		File: filePath,
	}

	result, err := p.Parse(context.Background(), req)

	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// Check that Unicode content is preserved
	if !strings.Contains(result.Text, "Hello 世界! 🌍") {
		t.Errorf("Parse() result missing Unicode content: %q", result.Text)
	}
	if !strings.Contains(result.Text, "Привет мир!") {
		t.Errorf("Parse() result missing Unicode content: %q", result.Text)
	}
}

func TestParseBooleanAndNull(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "booleans.json")

	// Create JSON with boolean and null values
	content := []byte(`{
  "active": true,
  "verified": false,
  "optional": null
}`)
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create boolean JSON test file: %v", err)
	}

	p := &Parser{}
	req := parser.ParseRequest{
		File: filePath,
	}

	result, err := p.Parse(context.Background(), req)

	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// Check that boolean and null values are present
	if !strings.Contains(result.Text, "active: true") {
		t.Errorf("Parse() result missing 'active: true': %q", result.Text)
	}
	if !strings.Contains(result.Text, "verified: false") {
		t.Errorf("Parse() result missing 'verified: false': %q", result.Text)
	}
	if !strings.Contains(result.Text, "optional: null") {
		t.Errorf("Parse() result missing 'optional: null': %q", result.Text)
	}
}

func TestParseNumbers(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "numbers.json")

	// Create JSON with various number types
	content := []byte(`{
  "age": 19,
  "price": 29.99,
  "quantity": 0,
  "temperature": -5.5
}`)
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create numbers JSON test file: %v", err)
	}

	p := &Parser{}
	req := parser.ParseRequest{
		File: filePath,
	}

	result, err := p.Parse(context.Background(), req)

	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// Check that numbers are present
	if !strings.Contains(result.Text, "age: 19") {
		t.Errorf("Parse() result missing 'age: 19': %q", result.Text)
	}
	if !strings.Contains(result.Text, "price: 29.99") {
		t.Errorf("Parse() result missing 'price: 29.99': %q", result.Text)
	}
}

func TestErrorWrapping(t *testing.T) {
	p := &Parser{}
	req := parser.ParseRequest{
		File: "nonexistent.json",
	}

	_, err := p.Parse(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}

	// Check that error contains expected message
	if !containsError(err.Error(), "failed to read JSON file") {
		t.Errorf("Error message = %q, want to contain 'failed to read JSON file'", err.Error())
	}
}

// containsError checks if a string contains a substring.
func containsError(s, substr string) bool {
	return strings.Contains(s, substr)
}

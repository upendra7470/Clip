package odt

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/upendra7470/clip/internal/filetype"
	"github.com/upendra7470/clip/internal/parser"
)

func TestFileType(t *testing.T) {
	p := &Parser{}
	if p.FileType() != filetype.FileTypeODT {
		t.Errorf("Expected FileType ODT, got %v", p.FileType())
	}
}

func TestParseMissingFile(t *testing.T) {
	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: "nonexistent.odt"})
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestParseInvalidZIP(t *testing.T) {
	// Create a temporary file with invalid content
	tmpFile := createTempFile(t, []byte("not a zip file"))
	defer os.Remove(tmpFile)

	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err == nil {
		t.Error("Expected error for invalid ZIP, got nil")
	}
}

func TestParseMissingContentXML(t *testing.T) {
	// Create a ZIP file without content.xml
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	// Add a different file instead of content.xml
	f, _ := zipWriter.Create("wrong.xml")
	f.Write([]byte("<root/>"))
	zipWriter.Close()

	tmpFile := createTempFile(t, buf.Bytes())
	defer os.Remove(tmpFile)

	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err == nil {
		t.Error("Expected error for missing content.xml, got nil")
	}
}

func TestParseInvalidXML(t *testing.T) {
	// Create a ZIP file with invalid XML in content.xml
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	f, _ := zipWriter.Create("content.xml")
	f.Write([]byte("not valid xml"))
	zipWriter.Close()

	tmpFile := createTempFile(t, buf.Bytes())
	defer os.Remove(tmpFile)

	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err == nil {
		t.Error("Expected error for invalid XML, got nil")
	}
}

func TestParseEmptyDocument(t *testing.T) {
	// Create a ZIP file with empty content.xml
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	f, _ := zipWriter.Create("content.xml")
	f.Write([]byte(`<?xml version="1.0"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
                         xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:text>
    </office:text>
  </office:body>
</office:document-content>`))
	zipWriter.Close()

	tmpFile := createTempFile(t, buf.Bytes())
	defer os.Remove(tmpFile)

	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err == nil {
		t.Error("Expected error for empty document, got nil")
	}
}

func TestParseSimpleText(t *testing.T) {
	// Create a ZIP file with simple text content
	content := `<?xml version="1.0"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
                         xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:text>
      <text:p>Hello World</text:p>
    </office:text>
  </office:body>
</office:document-content>`

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	f, _ := zipWriter.Create("content.xml")
	f.Write([]byte(content))
	zipWriter.Close()

	tmpFile := createTempFile(t, buf.Bytes())
	defer os.Remove(tmpFile)

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Hello World"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseMultipleParagraphs(t *testing.T) {
	// Create a ZIP file with multiple paragraphs
	content := `<?xml version="1.0"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
                         xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:text>
      <text:p>First paragraph</text:p>
      <text:p>Second paragraph</text:p>
      <text:p>Third paragraph</text:p>
    </office:text>
  </office:body>
</office:document-content>`

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	f, _ := zipWriter.Create("content.xml")
	f.Write([]byte(content))
	zipWriter.Close()

	tmpFile := createTempFile(t, buf.Bytes())
	defer os.Remove(tmpFile)

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "First paragraph\nSecond paragraph\nThird paragraph"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseNestedElements(t *testing.T) {
	// Create a ZIP file with nested elements
	content := `<?xml version="1.0"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
                         xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:text>
      <text:p>Paragraph with <text:span>nested</text:span> elements</text:p>
    </office:text>
  </office:body>
</office:document-content>`

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	f, _ := zipWriter.Create("content.xml")
	f.Write([]byte(content))
	zipWriter.Close()

	tmpFile := createTempFile(t, buf.Bytes())
	defer os.Remove(tmpFile)

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Paragraph with nested elements"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseUnicodeContent(t *testing.T) {
	// Create a ZIP file with Unicode content
	content := `<?xml version="1.0"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
                         xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:text>
      <text:p>Hello 世界 🌍 Привет</text:p>
    </office:text>
  </office:body>
</office:document-content>`

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	f, _ := zipWriter.Create("content.xml")
	f.Write([]byte(content))
	zipWriter.Close()

	tmpFile := createTempFile(t, buf.Bytes())
	defer os.Remove(tmpFile)

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Hello 世界 🌍 Привет"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseWhitespaceHandling(t *testing.T) {
	// Create a ZIP file with whitespace
	content := `<?xml version="1.0"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
                         xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body>
    <office:text>
      <text:p>  Leading and trailing spaces  </text:p>
      <text:p>Multiple   spaces   between   words</text:p>
    </office:text>
  </office:body>
</office:document-content>`

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	f, _ := zipWriter.Create("content.xml")
	f.Write([]byte(content))
	zipWriter.Close()

	tmpFile := createTempFile(t, buf.Bytes())
	defer os.Remove(tmpFile)

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Leading and trailing spaces\nMultiple   spaces   between   words"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestErrorWrapping(t *testing.T) {
	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: "nonexistent.odt"})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Test error unwrapping
	var odtErr *ODTParserError
	if !errors.As(err, &odtErr) {
		t.Error("Error should be unwrappable to *ODTParserError")
	}
}

// Helper function to create temporary files
func createTempFile(t *testing.T, content []byte) string {
	tmpFile, err := os.CreateTemp("", "test*.odt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	_, err = tmpFile.Write(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	return tmpFile.Name()
}

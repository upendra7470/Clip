package html

import (
	"context"
	"os"
	"testing"

	"github.com/upendra7470/clip/internal/parser"
)

func TestFileType(t *testing.T) {
	p := &Parser{}
	fileType := p.FileType()
	if string(fileType) != "HTML" {
		t.Errorf("Expected file type HTML, got %s", fileType)
	}
}

func TestParseMissingFile(t *testing.T) {
	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: "nonexistent.html"})
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestParseEmptyHTML(t *testing.T) {
	// Create temporary empty HTML file
	tmpFile, err := os.CreateTemp("", "empty*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Close the file immediately to ensure it's empty
	tmpFile.Close()

	p := &Parser{}
	_, err = p.Parse(context.Background(), parser.ParseRequest{File: tmpFile.Name()})
	if err == nil {
		t.Error("Expected error for empty HTML file, got nil")
	}
}

func TestParseSimpleHTML(t *testing.T) {
	// Create temporary HTML file with simple content
	tmpFile, err := os.CreateTemp("", "simple*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	htmlContent := `<html>
<body>
<h1>Hello World</h1>
<p>This is Clip.</p>
</body>
</html>`

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Hello World\nThis is Clip."
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseNestedHTML(t *testing.T) {
	// Create temporary HTML file with nested content
	tmpFile, err := os.CreateTemp("", "nested*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	htmlContent := `<div>
  <section>
    <h2>Title</h2>
    <p>Description</p>
  </section>
</div>`

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Title\nDescription"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseWithScriptsAndStyles(t *testing.T) {
	// Create temporary HTML file with scripts and styles
	tmpFile, err := os.CreateTemp("", "withscripts*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	htmlContent := `<html>
<head>
<style>
body { color:red; }
</style>
<script>
console.log("test");
</script>
</head>
<body>
<h1>Hello</h1>
</body>
</html>`

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Hello"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseWithComments(t *testing.T) {
	// Create temporary HTML file with comments
	tmpFile, err := os.CreateTemp("", "withcomments*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	htmlContent := `<!-- This is a comment -->
<h1>Main Title</h1>
<!-- Another comment -->
<p>Main content</p>`

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Main Title\nMain content"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseWithAttributes(t *testing.T) {
	// Create temporary HTML file with attributes
	tmpFile, err := os.CreateTemp("", "withattrs*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	htmlContent := `<div class="container" id="main">
<h1 class="title" data-test="value">Heading</h1>
<p class="content">Paragraph text</p>
</div>`

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Heading\nParagraph text"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseUnicodeContent(t *testing.T) {
	// Create temporary HTML file with unicode content
	tmpFile, err := os.CreateTemp("", "unicode*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	htmlContent := `<html>
<body>
<h1>Hello 世界</h1>
<p>This is a test with unicode: café, naïve, 日本語</p>
</body>
</html>`

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "Hello 世界\nThis is a test with unicode: café, naïve, 日本語"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseWhitespaceHandling(t *testing.T) {
	// Create temporary HTML file with various whitespace
	tmpFile, err := os.CreateTemp("", "whitespace*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	htmlContent := `<div>
    <p>First line</p>
    <p>Second line</p>
    <p>Third line</p>
</div>`

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "First line\nSecond line\nThird line"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseMultipleTextNodes(t *testing.T) {
	// Create temporary HTML file with multiple text nodes
	tmpFile, err := os.CreateTemp("", "multinodes*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	htmlContent := `<p>First part <strong>bold text</strong> second part</p>`

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "First part bold text second part"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestHTMLExtension(t *testing.T) {
	// Test .htm extension as well
	tmpFile, err := os.CreateTemp("", "test*.htm")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	htmlContent := `<html><body><h1>HTM Test</h1></body></html>`

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile.Name()})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "HTM Test"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

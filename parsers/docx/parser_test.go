package docx

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upendra7470/clip/internal/parser"
)

func TestParseRange(t *testing.T) {
	// Create a temporary test DOCX file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.docx")
	createTestDOCX(t, testFile, "Paragraph 1\nParagraph 2\nParagraph 3\nParagraph 4\nParagraph 5")

	// Test requesting 2-4 from a document with at least 4 paragraphs
	docxParser := &Parser{}
	req := parser.ParseRequest{
		File: testFile,
	}
	result, err := docxParser.ParseRange(context.Background(), req, 2, 4)
	if err != nil {
		t.Fatalf("ParseRange failed: %v", err)
	}
	assert.NoError(t, err)
	assert.NotContains(t, result.Text, "Warning: Requested range")
	assert.Contains(t, result.Text, "Paragraph 2")
	assert.Contains(t, result.Text, "Paragraph 3")
	assert.Contains(t, result.Text, "Paragraph 4")

	// Test requesting a range that exceeds document length
	result, err = docxParser.ParseRange(context.Background(), req, 4, 10)
	assert.NoError(t, err)
	assert.NotContains(t, result.Text, "Warning: Requested range")
	assert.Contains(t, result.Text, "Paragraph 4")
	assert.Contains(t, result.Text, "Paragraph 5")

	// Test clipboard content contains only extracted document content
	result, err = docxParser.ParseRange(context.Background(), req, 1, 3)
	assert.NoError(t, err)
	assert.NotContains(t, result.Text, "Warning: Requested range")
	assert.Contains(t, result.Text, "Paragraph 1")
	assert.Contains(t, result.Text, "Paragraph 2")
	assert.Contains(t, result.Text, "Paragraph 3")
	assert.NotContains(t, result.Text, "Paragraph 4")
	assert.NotContains(t, result.Text, "Paragraph 5")
}

func createTestDOCX(t *testing.T, path string, content string) {
	// Create a minimal but valid DOCX file (ZIP archive with word/document.xml)
	// This creates the minimal structure that the DOCX parser expects

	// Create a temporary directory for the DOCX contents
	tempDir := t.TempDir()

	// Create word/document.xml with the test content wrapped in proper XML
	documentXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:body>`

	// Split content by lines and create paragraphs
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		documentXML += `
		<w:p>
			<w:r>
				<w:t>` + line + `</w:t>
			</w:r>
		</w:p>`
	}

	documentXML += `
	</w:body>
</w:document>`

	// Write document.xml
	documentXMLPath := filepath.Join(tempDir, "word", "document.xml")
	err := os.MkdirAll(filepath.Dir(documentXMLPath), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(documentXMLPath, []byte(documentXML), 0644)
	assert.NoError(t, err)

	// Create the DOCX file (ZIP archive)
	zipFile, err := os.Create(path)
	assert.NoError(t, err)
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add document.xml to the ZIP archive
	xmlFile, err := zipWriter.Create("word/document.xml")
	assert.NoError(t, err)

	_, err = xmlFile.Write([]byte(documentXML))
	assert.NoError(t, err)
}

package docx

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func createTestDOCXFromXML(t *testing.T, path string, xmlContent string) {
	tempDir := t.TempDir()
	dst := filepath.Join(tempDir, "word", "document.xml")

	// Create directory structure
	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		t.Fatalf("Failed to create directory structure: %v", err)
	}

	// Write the document.xml with provided content
	err = os.WriteFile(dst, []byte(xmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write document.xml: %v", err)
	}

	// Create the DOCX file (ZIP archive)
	zipFile, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create DOCX file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add document.xml to the ZIP archive
	xmlFile, err := zipWriter.Create("word/document.xml")
	if err != nil {
		t.Fatalf("Failed to create XML file in ZIP: %v", err)
	}

	_, err = xmlFile.Write([]byte(xmlContent))
	if err != nil {
		t.Fatalf("Failed to write XML content to ZIP: %v", err)
	}
}

func extractContentFromXML(xmlContent string) (string, error) {
	text, _, err := extractStructuredContentFromXML(xmlContent, false)
	return text, err
}

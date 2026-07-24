package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// createTestDOCX creates a DOCX file with the provided test content
func createTestDOCX(path string, content string) error {
	// Create a temporary directory for the DOCX structure
	tempDir := filepath.Join(".", "test_docx_temp")
	os.RemoveAll(tempDir) // Clean up any existing temp directory

	// Create directory structure
	err := os.MkdirAll(filepath.Join(tempDir, "word"), 0755)
	if err != nil {
		return err
	}

	// Create the document.xml with test content
	testContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:body>`
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		testContent += `
		<w:p>
			<w:r>
				<w:t>` + line + `</w:t>
			</w:r>
		</w:p>`
	}
	testContent += `
	</w:body>
</w:document>`
	srcContent := []byte(testContent)

	// Write the document.xml to the temporary location
	err = os.WriteFile(filepath.Join(tempDir, "word", "document.xml"), srcContent, 0644)
	if err != nil {
		return err
	}

	// Create the DOCX file (ZIP archive)
	zipFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add document.xml to the ZIP archive
	xmlFile, err := zipWriter.Create("word/document.xml")
	if err != nil {
		return err
	}

	_, err = xmlFile.Write(srcContent)
	if err != nil {
		return err
	}

	// Clean up temp directory
	os.RemoveAll(tempDir)

	return nil
}

func main() {
	// Create a test DOCX file with known content
	testContent := `Paragraph 1: This is the first paragraph of the test document.
Paragraph 2: This is the second paragraph with some additional content.
Paragraph 3: This paragraph contains more detailed information for testing.
Paragraph 4: Here we have the fourth paragraph to test range extraction.
Paragraph 5: This is the fifth and final paragraph in this test document.`

	err := createTestDOCX("test_range.docx", testContent)
	if err != nil {
		fmt.Printf("Error creating test DOCX file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Created test_range.docx successfully")
	fmt.Println("Content:")
	fmt.Println(testContent)
}

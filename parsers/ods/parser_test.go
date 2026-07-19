package ods

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

// createTempFile creates a temporary file with the given content and returns its path.
func createTempFile(t *testing.T, content []byte) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test*.ods")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	return tmpFile.Name()
}

// createTempODSFile creates a temporary ODS file with the given content.xml content.
func createTempODSFile(t *testing.T, contentXML string) string {
	t.Helper()
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Create content.xml file in the ZIP archive
	f, err := zipWriter.Create("content.xml")
	if err != nil {
		t.Fatalf("Failed to create content.xml in ZIP: %v", err)
	}

	if _, err := f.Write([]byte(contentXML)); err != nil {
		t.Fatalf("Failed to write content.xml: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close ZIP writer: %v", err)
	}

	// Create temporary file
	tmpFile := createTempFile(t, buf.Bytes())
	return tmpFile
}

func TestFileType(t *testing.T) {
	p := &Parser{}
	if p.FileType() != filetype.FileTypeODS {
		t.Errorf("Expected FileType ODS, got %v", p.FileType())
	}
}

func TestParseMissingFile(t *testing.T) {
	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: "nonexistent.ods"})
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
	// Create a ZIP file with invalid XML content
	invalidXML := "This is not valid XML content"
	tmpFile := createTempODSFile(t, invalidXML)
	defer os.Remove(tmpFile)

	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err == nil {
		t.Error("Expected error for invalid XML, got nil")
	}
}

func TestParseEmptySpreadsheet(t *testing.T) {
	// Create a valid ODS with empty spreadsheet
	emptyODS := `<?xml version="1.0" encoding="UTF-8"?>
	<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
		<office:body>
			<office:spreadsheet>
				<table:table table:name="Sheet1">
					<table:table-row/>
				</table:table>
			</office:spreadsheet>
		</office:body>
	</office:document-content>`

	tmpFile := createTempODSFile(t, emptyODS)
	defer os.Remove(tmpFile)

	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err == nil {
		t.Error("Expected error for empty spreadsheet, got nil")
	}
}

func TestParseSimpleSpreadsheet(t *testing.T) {
	// Create a simple ODS with one cell containing text
	simpleODS := `<?xml version="1.0" encoding="UTF-8"?>
	<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
		<office:body>
			<office:spreadsheet>
				<table:table table:name="Sheet1">
					<table:table-row>
						<table:table-cell>
							<text:p>Hello World</text:p>
						</table:table-cell>
					</table:table-row>
				</table:table>
			</office:spreadsheet>
		</office:body>
	</office:document-content>`

	tmpFile := createTempODSFile(t, simpleODS)
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

func TestParseMultipleRows(t *testing.T) {
	// Create ODS with multiple rows
	multiRowODS := `<?xml version="1.0" encoding="UTF-8"?>
	<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
		<office:body>
			<office:spreadsheet>
				<table:table table:name="Sheet1">
					<table:table-row>
						<table:table-cell>
							<text:p>Row1Cell1</text:p>
						</table:table-cell>
					</table:table-row>
					<table:table-row>
						<table:table-cell>
							<text:p>Row2Cell1</text:p>
						</table:table-cell>
					</table:table-row>
				</table:table>
			</office:spreadsheet>
		</office:body>
	</office:document-content>`

	tmpFile := createTempODSFile(t, multiRowODS)
	defer os.Remove(tmpFile)

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Row1Cell1\nRow2Cell1"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseMultipleColumns(t *testing.T) {
	// Create ODS with multiple columns
	multiColODS := `<?xml version="1.0" encoding="UTF-8"?>
	<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
		<office:body>
			<office:spreadsheet>
				<table:table table:name="Sheet1">
					<table:table-row>
						<table:table-cell>
							<text:p>Name</text:p>
						</table:table-cell>
						<table:table-cell>
							<text:p>Age</text:p>
						</table:table-cell>
					</table:table-row>
					<table:table-row>
						<table:table-cell>
							<text:p>Sai</text:p>
						</table:table-cell>
						<table:table-cell>
							<text:p>19</text:p>
						</table:table-cell>
					</table:table-row>
				</table:table>
			</office:spreadsheet>
		</office:body>
	</office:document-content>`

	tmpFile := createTempODSFile(t, multiColODS)
	defer os.Remove(tmpFile)

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Name\nAge\nSai\n19"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseUnicode(t *testing.T) {
	// Create ODS with Unicode characters
	unicodeODS := `<?xml version="1.0" encoding="UTF-8"?>
	<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
		<office:body>
			<office:spreadsheet>
				<table:table table:name="Sheet1">
					<table:table-row>
						<table:table-cell>
							<text:p>Hello 世界</text:p>
						</table:table-cell>
					</table:table-row>
					<table:table-row>
						<table:table-cell>
							<text:p>Привет</text:p>
						</table:table-cell>
					</table:table-row>
				</table:table>
			</office:spreadsheet>
		</office:body>
	</office:document-content>`

	tmpFile := createTempODSFile(t, unicodeODS)
	defer os.Remove(tmpFile)

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Hello 世界\nПривет"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseEmptyCells(t *testing.T) {
	// Create ODS with empty cells
	emptyCellsODS := `<?xml version="1.0" encoding="UTF-8"?>
	<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
		<office:body>
			<office:spreadsheet>
				<table:table table:name="Sheet1">
					<table:table-row>
						<table:table-cell>
							<text:p>Cell1</text:p>
						</table:table-cell>
						<table:table-cell>
							<text:p></text:p>
						</table:table-cell>
						<table:table-cell>
							<text:p>Cell3</text:p>
						</table:table-cell>
					</table:table-row>
				</table:table>
			</office:spreadsheet>
		</office:body>
	</office:document-content>`

	tmpFile := createTempODSFile(t, emptyCellsODS)
	defer os.Remove(tmpFile)

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Cell1\nCell3"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseNestedXML(t *testing.T) {
	// Create ODS with nested XML structure
	nestedODS := `<?xml version="1.0" encoding="UTF-8"?>
	<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
		<office:body>
			<office:spreadsheet>
				<table:table table:name="Sheet1">
					<table:table-row>
						<table:table-cell>
							<text:p>Text with <text:span>formatting</text:span> inside</text:p>
						</table:table-cell>
					</table:table-row>
				</table:table>
			</office:spreadsheet>
		</office:body>
	</office:document-content>`

	tmpFile := createTempODSFile(t, nestedODS)
	defer os.Remove(tmpFile)

	p := &Parser{}
	result, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should extract all text content including nested elements
	expected := "Text with formatting inside"
	if result.Text != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Text)
	}
}

func TestParseErrorWrapping(t *testing.T) {
	// Test that errors are properly wrapped
	tmpFile := createTempFile(t, []byte("invalid content"))
	defer os.Remove(tmpFile)

	p := &Parser{}
	_, err := p.Parse(context.Background(), parser.ParseRequest{File: tmpFile})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check that error can be unwrapped
	var odsErr *ODSParserError
	if !errors.As(err, &odsErr) {
		t.Error("Error is not an ODSParserError")
	}

	if odsErr.Error() == "" {
		t.Error("Error message is empty")
	}
}

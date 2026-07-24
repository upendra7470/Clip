package docx

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractStructuredContentFromXML_Tables(t *testing.T) {
	testXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:body>
        <w:p>
            <w:r>
                <w:t>Paragraph before table</w:t>
            </w:r>
        </w:p>
        <w:tbl>
            <w:tr>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Header 1</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Header 2</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
            </w:tr>
            <w:tr>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Row 1, Cell 1</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Row 1, Cell 2</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
            </w:tr>
            <w:tr>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Row 2, Cell 1</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Row 2, Cell 2</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
            </w:tr>
        </w:tbl>
        <w:p>
            <w:r>
                <w:t>Paragraph after table</w:t>
            </w:r>
        </w:p>
    </w:body>
</w:document>`

	result, _, err := extractStructuredContentFromXML(testXML, false)
	assert.NoError(t, err)

	// Verify paragraphs
	assert.Contains(t, result, "Paragraph before table")
	assert.Contains(t, result, "Paragraph after table")

	// Verify table structure
	assert.Contains(t, result, "| Header 1 | Header 2 |")
	assert.Contains(t, result, "| Row 1, Cell 1 | Row 1, Cell 2 |")
	assert.Contains(t, result, "| Row 2, Cell 1 | Row 2, Cell 2 |")

	// Verify table separator
	assert.Contains(t, result, "| --- | --- |")
}

func TestExtractStructuredContentFromXML_NestedParagraphsInTableCells(t *testing.T) {
	testXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:body>
        <w:p>
            <w:r>
                <w:t>Paragraph before table</w:t>
            </w:r>
        </w:p>
        <w:tbl>
            <w:tr>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Cell with nested paragraphs</w:t>
                        </w:r>
                    </w:p>
                    <w:p>
                        <w:r>
                            <w:t>First nested paragraph</w:t>
                        </w:r>
                    </w:p>
                    <w:p>
                        <w:r>
                            <w:t>Second nested paragraph</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Simple cell</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
            </w:tr>
        </w:tbl>
        <w:p>
            <w:r>
                <w:t>Paragraph after table</w:t>
            </w:r>
        </w:p>
    </w:body>
</w:document>`

	result, _, err := extractStructuredContentFromXML(testXML, false)
	assert.NoError(t, err)

	// Verify paragraphs
	assert.Contains(t, result, "Paragraph before table")
	assert.Contains(t, result, "Paragraph after table")

	// Verify nested paragraphs in table cells are handled
	assert.Contains(t, result, "Cell with nested paragraphs")
	assert.Contains(t, result, "First nested paragraph")
	assert.Contains(t, result, "Second nested paragraph")
	assert.Contains(t, result, "Simple cell")
}

func TestExtractStructuredContentFromXML_UnicodePreservation(t *testing.T) {
	testXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:body>
        <w:p>
            <w:r>
                <w:t>English paragraph</w:t>
            </w:r>
        </w:p>
        <w:p>
            <w:r>
                <w:t>段落中文</w:t>
            </w:r>
        </w:p>
        <w:p>
            <w:r>
                <w:t>Параграф на русском</w:t>
            </w:r>
        </w:p>
        <w:p>
            <w:r>
                <w:t>paragraph avec des caractères spéciaux: é, è, ê</w:t>
            </w:r>
        </w:p>
    </w:body>
</w:document>`

	result, _, err := extractStructuredContentFromXML(testXML, false)
	assert.NoError(t, err)

	// Verify Unicode preservation
	assert.Contains(t, result, "English paragraph")
	assert.Contains(t, result, "段落中文")
	assert.Contains(t, result, "Параграф на русском")
	assert.Contains(t, result, "paragraph avec des caractères spéciaux: é, è, ê")
}

func TestExtractStructuredContentFromXML_MixedDocumentContent(t *testing.T) {
	testXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:body>
        <w:p>
            <w:r>
                <w:t>First paragraph</w:t>
            </w:r>
        </w:p>
        <w:tbl>
            <w:tr>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Table Cell 1</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Table Cell 2</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
            </w:tr>
        </w:tbl>
        <w:p>
            <w:r>
                <w:t>Paragraph after table</w:t>
            </w:r>
        </w:p>
        <w:p>
            <w:r>
                <w:t>段落中文</w:t>
            </w:r>
        </w:p>
        <w:tbl>
            <w:tr>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Nested paragraph 1</w:t>
                        </w:r>
                    </w:p>
                    <w:p>
                        <w:r>
                            <w:t>Nested paragraph 2</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Simple cell</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
            </w:tr>
        </w:tbl>
        <w:p>
            <w:r>
                <w:t>Final paragraph</w:t>
            </w:r>
        </w:p>
    </w:body>
</w:document>`

	result, _, err := extractStructuredContentFromXML(testXML, false)
	assert.NoError(t, err)

	// Verify mixed content
	assert.Contains(t, result, "First paragraph")
	assert.Contains(t, result, "Table Cell 1")
	assert.Contains(t, result, "Table Cell 2")
	assert.Contains(t, result, "Paragraph after table")
	assert.Contains(t, result, "段落中文")
	assert.Contains(t, result, "Nested paragraph 1")
	assert.Contains(t, result, "Nested paragraph 2")
	assert.Contains(t, result, "Simple cell")
	assert.Contains(t, result, "Final paragraph")

	// Verify table structure
	assert.Contains(t, result, "| Table Cell 1 | Table Cell 2 |")
	assert.Contains(t, result, "| Nested paragraph 1")
	assert.Contains(t, result, "| Simple cell |")
}

func TestExtractStructuredContentFromXML_EmptyDocument(t *testing.T) {
	testXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:body>
    </w:body>
</w:document>`

	result, _, err := extractStructuredContentFromXML(testXML, false)
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestExtractStructuredContentFromXML_OnlyParagraphs(t *testing.T) {
	testXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:body>
        <w:p>
            <w:r>
                <w:t>First paragraph</w:t>
            </w:r>
        </w:p>
        <w:p>
            <w:r>
                <w:t>Second paragraph</w:t>
            </w:r>
        </w:p>
        <w:p>
            <w:r>
                <w:t>Third paragraph</w:t>
            </w:r>
        </w:p>
    </w:body>
</w:document>`

	result, _, err := extractStructuredContentFromXML(testXML, false)
	assert.NoError(t, err)

	// Verify paragraphs are separated by newlines
	lines := strings.Split(strings.TrimSpace(result), "\n")
	assert.Equal(t, 3, len(lines))
	assert.Contains(t, result, "First paragraph")
	assert.Contains(t, result, "Second paragraph")
	assert.Contains(t, result, "Third paragraph")
}

func TestExtractStructuredContentFromXML_OnlyTables(t *testing.T) {
	testXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:body>
        <w:tbl>
            <w:tr>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Cell 1</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Cell 2</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
            </w:tr>
            <w:tr>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Cell 3</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
                <w:tc>
                    <w:p>
                        <w:r>
                            <w:t>Cell 4</w:t>
                        </w:r>
                    </w:p>
                </w:tc>
            </w:tr>
        </w:tbl>
    </w:body>
</w:document>`

	result, _, err := extractStructuredContentFromXML(testXML, false)
	assert.NoError(t, err)

	// Verify table structure
	assert.Contains(t, result, "| Cell 1 | Cell 2 |")
	assert.Contains(t, result, "| Cell 3 | Cell 4 |")
	assert.Contains(t, result, "| --- | --- |")
}

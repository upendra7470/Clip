package docx

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upendra7470/clip/internal/parser"
)

func TestExtractBlocks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Full extraction",
			content:  "Test content",
			start:    1,
			end:      1,
			expected: "Test content",
		},
		{
			name:     "Block range extraction",
			content:  "Block 1\nBlock 2\nBlock 3",
			start:    2,
			end:      2,
			expected: "Block 2",
		},
		{
			name:     "Out of range block range",
			content:  "Block 1\nBlock 2\nBlock 3",
			start:    4,
			end:      5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result, err := parser.ExtractBlocks(tt.content, tt.start, tt.end)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseRangeWithTables tests that ParseRange preserves table structure
func TestParseRangeWithTables(t *testing.T) {
	// Create a test DOCX file with paragraphs and tables
	xmlContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
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
	               <w:t>Third paragraph</w:t>
	           </w:r>
	       </w:p>
	   </w:body>
</w:document>`

	// Create temporary DOCX file
	tempDir := t.TempDir()
	docxPath := filepath.Join(tempDir, "test.docx")
	createTestDOCXFromXML(t, docxPath, xmlContent)

	docParser := NewParser()

	// Test 1: Parse full document should include table
	req := parser.ParseRequest{File: docxPath}
	result, err := docParser.Parse(context.Background(), req)
	assert.NoError(t, err)
	assert.Contains(t, result.Text, "First paragraph")
	assert.Contains(t, result.Text, "Second paragraph")
	assert.Contains(t, result.Text, "| Table Cell 1 | Table Cell 2 |")
	assert.Contains(t, result.Text, "Third paragraph")

	// Test 2: ParseRange should use same structured path and preserve tables
	// Range 1-2 should get first two paragraphs
	result, err = docParser.ParseRange(context.Background(), req, 1, 2)
	assert.NoError(t, err)
	assert.Contains(t, result.Text, "First paragraph")
	assert.Contains(t, result.Text, "Second paragraph")
	// Should NOT contain table or third paragraph
	assert.NotContains(t, result.Text, "Table Cell 1")
	assert.NotContains(t, result.Text, "Third paragraph")

	// Test 3: ParseRange with range that includes table (paragraphs 3-4)
	// The table counts as one paragraph unit, and third paragraph as another
	result, err = docParser.ParseRange(context.Background(), req, 3, 4)
	assert.NoError(t, err)
	// Should contain table and third paragraph
	assert.Contains(t, result.Text, "| Table Cell 1 | Table Cell 2 |")
	assert.Contains(t, result.Text, "Third paragraph")
	// Should NOT contain first paragraphs
	assert.NotContains(t, result.Text, "First paragraph")
	assert.NotContains(t, result.Text, "Second paragraph")
}

// TestParseRangeConsistency tests that Parse and ParseRange use the same parsing logic
func TestParseRangeConsistency(t *testing.T) {
	// Create a test DOCX file with complex content
	xmlContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	   <w:body>
	       <w:p>
	           <w:r>
	               <w:t>Paragraph 1</w:t>
	           </w:r>
	       </w:p>
	       <w:tbl>
	           <w:tr>
	               <w:tc>
	                   <w:p>
	                       <w:r>
	                           <w:t>Header</w:t>
	                       </w:r>
	                   </w:p>
	               </w:tc>
	           </w:tr>
	           <w:tr>
	               <w:tc>
	                   <w:p>
	                       <w:r>
	                           <w:t>Data</w:t>
	                       </w:r>
	                   </w:p>
	               </w:tc>
	           </w:tr>
	       </w:tbl>
	       <w:p>
	           <w:r>
	               <w:t>Paragraph 2</w:t>
	           </w:r>
	       </w:p>
	   </w:body>
</w:document>`

	// Create temporary DOCX file
	tempDir := t.TempDir()
	docxPath := filepath.Join(tempDir, "test.docx")
	createTestDOCXFromXML(t, docxPath, xmlContent)

	docParser := NewParser()
	req := parser.ParseRequest{File: docxPath}

	// Parse full document
	fullResult, err := docParser.Parse(context.Background(), req)
	assert.NoError(t, err)

	// Parse full range - get total paragraphs first by parsing
	// We'll use a large number for end to get all content
	rangeResult, err := docParser.ParseRange(context.Background(), req, 1, 100)
	assert.NoError(t, err)

	// Both should contain the same content (tables and paragraphs)
	// Check that both contain the essential elements
	assert.Contains(t, fullResult.Text, "Paragraph 1")
	assert.Contains(t, fullResult.Text, "| Header |")
	assert.Contains(t, fullResult.Text, "| Data |")
	assert.Contains(t, fullResult.Text, "Paragraph 2")

	assert.Contains(t, rangeResult.Text, "Paragraph 1")
	assert.Contains(t, rangeResult.Text, "| Header |")
	assert.Contains(t, rangeResult.Text, "| Data |")
	assert.Contains(t, rangeResult.Text, "Paragraph 2")
}

// TestParseRangePreservesUnicode tests that ParseRange preserves Unicode characters
func TestParseRangePreservesUnicode(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	   <w:body>
	       <w:p>
	           <w:r>
	               <w:t>English text</w:t>
	           </w:r>
	       </w:p>
	       <w:p>
	           <w:r>
	               <w:t>中文文字</w:t>
	           </w:r>
	       </w:p>
	       <w:p>
	           <w:r>
	               <w:t>Русский текст</w:t>
	           </w:r>
	       </w:p>
	   </w:body>
</w:document>`

	// Create temporary DOCX file
	tempDir := t.TempDir()
	docxPath := filepath.Join(tempDir, "test.docx")
	createTestDOCXFromXML(t, docxPath, xmlContent)

	docParser := NewParser()
	req := parser.ParseRequest{File: docxPath}

	// Parse range for middle paragraph (Chinese)
	result, err := docParser.ParseRange(context.Background(), req, 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, "中文文字", result.Text)
}

func TestExtractTables(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Table in selected range",
			content:  "Table 1\nTable 2\nTable 3",
			start:    2,
			end:      2,
			expected: "Table 2",
		},
		{
			name:     "Nested table paragraphs",
			content:  "Table 1\nNested Table 1\nNested Table 2\nTable 2",
			start:    2,
			end:      3,
			expected: "Nested Table 1\nNested Table 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result, err := parser.ExtractTables(tt.content, tt.start, tt.end)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

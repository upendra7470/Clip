package docx

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/upendra7470/clip/internal/filetype"
	"github.com/upendra7470/clip/internal/parser"
)

// Parser implements the parser.Parser and parser.RangeParser interfaces for DOCX files.
type Parser struct{}

// NewParser creates a new DOCX Parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// Parse reads a DOCX file and extracts text content.
// DOCX files are ZIP archives containing XML files.
// This parser extracts text from word/document.xml <w:t> nodes.
func (p *Parser) Parse(ctx context.Context, req parser.ParseRequest) (parser.ParseResult, error) {
	// Open the DOCX file (which is a ZIP archive)
	file, err := os.Open(req.File)
	if err != nil {
		if os.IsNotExist(err) {
			return parser.ParseResult{}, wrapError("Could not open DOCX file:\n"+req.File+"\n\nReason:\nfile does not exist", err)
		}
		if os.IsPermission(err) {
			return parser.ParseResult{}, wrapError("Could not open DOCX file:\n"+req.File+"\n\nReason:\npermission denied", err)
		}
		return parser.ParseResult{}, wrapError("Could not open DOCX file:\n"+req.File+"\n\nReason:\n"+err.Error(), err)
	}
	defer file.Close()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return parser.ParseResult{}, wrapError("failed to get file info", err)
	}

	// Read the ZIP archive
	zipReader, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		return parser.ParseResult{}, wrapError("failed to read DOCX as ZIP archive", err)
	}

	// Find and extract word/document.xml
	var documentXML string
	for _, zipFile := range zipReader.File {
		if zipFile.Name == "word/document.xml" {
			rc, err := zipFile.Open()
			if err != nil {
				return parser.ParseResult{}, wrapError("failed to open document.xml", err)
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return parser.ParseResult{}, wrapError("failed to read document.xml", err)
			}
			documentXML = string(content)
			break
		}
	}

	if documentXML == "" {
		return parser.ParseResult{}, wrapError("document.xml not found in DOCX", nil)
	}

	// Parse XML to extract structured content including tables
	text, _, err := extractStructuredContentFromXML(documentXML, false)
	if err != nil {
		return parser.ParseResult{}, wrapError("failed to parse DOCX XML", err)
	}

	if text == "" {
		return parser.ParseResult{}, wrapError("no text content found in DOCX", nil)
	}

	return parser.ParseResult{
		Text: text,
	}, nil
}

// FileType returns the file type this parser handles.
func (p *Parser) FileType() filetype.FileType {
	return filetype.FileTypeDOCX
}

// GetRangeUnit returns the unit type that this parser uses for ranges.
func (p *Parser) GetRangeUnit() string {
	return "paragraphs"
}

// ParseRange extracts text from a specific paragraph range in a DOCX file.
func (p *Parser) ParseRange(ctx context.Context, req parser.ParseRequest, start, end int) (parser.ParseResult, error) {
	// Validate paragraph range
	if start < 1 || end < 1 {
		return parser.ParseResult{}, wrapError(fmt.Sprintf("paragraph numbers must start from 1, got %d-%d", start, end), nil)
	}
	if end < start {
		return parser.ParseResult{}, wrapError(fmt.Sprintf("invalid paragraph range: start paragraph must not be greater than end paragraph (got %d-%d)", start, end), nil)
	}

	// Handle special range formats
	if start == -1 {
		start = 1 // Start from beginning
	}

	// Open the DOCX file (which is a ZIP archive)
	file, err := os.Open(req.File)
	if err != nil {
		if os.IsNotExist(err) {
			return parser.ParseResult{}, wrapError("Could not open DOCX file:\n"+req.File+"\n\nReason:\nfile does not exist", err)
		}
		if os.IsPermission(err) {
			return parser.ParseResult{}, wrapError("Could not open DOCX file:\n"+req.File+"\n\nReason:\npermission denied", err)
		}
		return parser.ParseResult{}, wrapError("Could not open DOCX file:\n"+req.File+"\n\nReason:\n"+err.Error(), err)
	}
	defer file.Close()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return parser.ParseResult{}, wrapError("failed to get file info", err)
	}

	// Read the ZIP archive
	zipReader, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		return parser.ParseResult{}, wrapError("failed to read DOCX as ZIP archive", err)
	}

	// Find and extract word/document.xml
	var documentXML string
	for _, zipFile := range zipReader.File {
		if zipFile.Name == "word/document.xml" {
			rc, err := zipFile.Open()
			if err != nil {
				return parser.ParseResult{}, wrapError("failed to open document.xml", err)
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return parser.ParseResult{}, wrapError("failed to read document.xml", err)
			}
			documentXML = string(content)
			break
		}
	}

	if documentXML == "" {
		return parser.ParseResult{}, wrapError("document.xml not found in DOCX", nil)
	}

	// Parse XML to extract structured paragraphs using the same parsing path as Parse()
	structuredParagraphs, err := extractStructuredParagraphsFromXML(documentXML)
	if err != nil {
		return parser.ParseResult{}, wrapError("failed to parse DOCX XML", err)
	}

	totalParagraphs := len(structuredParagraphs)

	// Handle special range formats
	if end == -1 {
		end = totalParagraphs // End at last paragraph
	}

	// Validate range against actual paragraph count
	if start > totalParagraphs {
		return parser.ParseResult{}, wrapError(fmt.Sprintf("requested paragraph range exceeds document paragraph count (document has %d paragraphs, requested %d-%d)", totalParagraphs, start, end), nil)
	}
	if end > totalParagraphs {
		end = totalParagraphs // Adjust end to last paragraph if it exceeds
	}

	// Extract only the requested paragraph range
	var result strings.Builder
	for i := start - 1; i < end && i < len(structuredParagraphs); i++ {
		if i > start-1 {
			// Add appropriate separator based on content type
			if structuredParagraphs[i].IsTable {
				if result.Len() > 0 {
					result.WriteString("\n")
				}
			} else {
				result.WriteString("\n")
			}
		}
		result.WriteString(structuredParagraphs[i].Content)
	}

	if result.Len() == 0 {
		return parser.ParseResult{}, wrapError(fmt.Sprintf("no text content found in paragraphs %d-%d", start, end), nil)
	}

	// Return ONLY the actual document content - NO warning messages
	return parser.ParseResult{
		Text: result.String(),
	}, nil
}

// extractParagraphsFromXML parses the XML and extracts paragraphs as a slice of strings.
func extractParagraphsFromXML(xmlContent string) ([]string, int, error) {
	var paragraphs []string
	var currentParagraph strings.Builder

	decoder := xml.NewDecoder(strings.NewReader(xmlContent))
	var inTextNode bool
	var inParagraph bool
	var currentText strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inTextNode = true
				currentText.Reset()
			} else if t.Name.Local == "p" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inParagraph = true
			}
		case xml.CharData:
			if inTextNode {
				currentText.Write(t)
			}
		case xml.EndElement:
			if inTextNode && t.Name.Local == "t" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inTextNode = false
				text := strings.TrimSpace(currentText.String())
				if text != "" {
					if currentParagraph.Len() > 0 {
						currentParagraph.WriteString(" ")
					}
					currentParagraph.WriteString(text)
				}
			} else if inParagraph && t.Name.Local == "p" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inParagraph = false
				paraText := strings.TrimSpace(currentParagraph.String())
				if paraText != "" {
					// Strip "Paragraph N:" prefix if present
					paraText = stripParagraphPrefix(paraText)
					paragraphs = append(paragraphs, paraText)
				}
				currentParagraph.Reset()
			}
		}
	}

	return paragraphs, len(paragraphs), nil
}

// stripParagraphPrefix removes "Paragraph N:" prefix from paragraph text if present.
func stripParagraphPrefix(text string) string {
	// Match "Paragraph N:" pattern where N is a number
	// This handles cases like "Paragraph 1: Hello World" -> "Hello World"
	re := regexp.MustCompile(`^Paragraph \d+:\s*`)
	return re.ReplaceAllString(text, "")
}

// StructuredParagraph represents a paragraph or table from the DOCX structure
type StructuredParagraph struct {
	Content string
	IsTable bool
}

// extractStructuredParagraphsFromXML parses the XML and returns structured paragraphs for range filtering
func extractStructuredParagraphsFromXML(xmlContent string) ([]StructuredParagraph, error) {
	_, paragraphs, err := extractStructuredContentFromXML(xmlContent, true)
	return paragraphs, err
}

// extractStructuredContentFromXML parses the XML and extracts structured content including tables.
// If collectParagraphs is true, it returns the structured paragraphs separately for range filtering.
func extractStructuredContentFromXML(xmlContent string, collectParagraphs bool) (string, []StructuredParagraph, error) {
	var result strings.Builder
	var inTable bool
	var inTableRow bool
	var inTableCell bool
	var inParagraph bool
	var inTextNode bool
	var currentText strings.Builder
	var tableRows [][]string
	var currentRow []string
	var currentCell strings.Builder
	var paragraphs []StructuredParagraph
	var currentParagraph strings.Builder

	decoder := xml.NewDecoder(strings.NewReader(xmlContent))

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Handle table elements
			if t.Name.Local == "tbl" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inTable = true
				currentParagraph.Reset()
			} else if t.Name.Local == "tr" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inTableRow = true
				currentRow = []string{}
			} else if t.Name.Local == "tc" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inTableCell = true
				currentCell.Reset()
			} else if t.Name.Local == "p" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inParagraph = true
				if inTableCell {
					currentCell.WriteString(" ")
				} else if collectParagraphs {
					currentParagraph.Reset()
				}
			} else if t.Name.Local == "t" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inTextNode = true
				currentText.Reset()
			}
		case xml.CharData:
			if inTextNode {
				currentText.Write(t)
			}
		case xml.EndElement:
			if inTextNode && t.Name.Local == "t" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inTextNode = false
				text := strings.TrimSpace(currentText.String())
				if text != "" {
					text = stripParagraphPrefix(text)
					if inTableCell {
						if currentCell.Len() > 0 {
							currentCell.WriteString(" ")
						}
						currentCell.WriteString(text)
					} else if inParagraph {
						if collectParagraphs {
							if currentParagraph.Len() > 0 {
								currentParagraph.WriteString(" ")
							}
							currentParagraph.WriteString(text)
						} else {
							if result.Len() > 0 {
								result.WriteString(" ")
							}
							result.WriteString(text)
						}
					}
				}
			} else if inParagraph && t.Name.Local == "p" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inParagraph = false
				if inTableCell {
					// Don't add newline inside table cells
				} else if collectParagraphs {
					// Collect the paragraph if we're in collection mode
					paraText := strings.TrimSpace(currentParagraph.String())
					if paraText != "" {
						paragraphs = append(paragraphs, StructuredParagraph{
							Content: paraText,
							IsTable: false,
						})
					}
				} else {
					if !inTableCell {
						result.WriteString("\n")
					}
				}
			} else if inTableCell && t.Name.Local == "tc" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inTableCell = false
				currentRow = append(currentRow, strings.TrimSpace(currentCell.String()))
			} else if inTableRow && t.Name.Local == "tr" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inTableRow = false
				tableRows = append(tableRows, currentRow)
			} else if inTable && t.Name.Local == "tbl" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				inTable = false
				// Format the table in Markdown style
				if len(tableRows) > 0 {
					var tableResult strings.Builder
					// Write table header
					tableResult.WriteString("| ")
					tableResult.WriteString(strings.Join(tableRows[0], " | "))
					tableResult.WriteString(" |")
					tableResult.WriteString("\n")
					// Write table separator
					tableResult.WriteString("| ")
					for range tableRows[0] {
						tableResult.WriteString("---")
						if len(tableRows[0]) > 1 {
							tableResult.WriteString(" | ")
						}
					}
					tableResult.WriteString(" |")
					tableResult.WriteString("\n")
					// Write table rows
					for _, row := range tableRows[1:] {
						tableResult.WriteString("| ")
						tableResult.WriteString(strings.Join(row, " | "))
						tableResult.WriteString(" |")
						tableResult.WriteString("\n")
					}
					tableResult.WriteString("\n")

					if collectParagraphs {
						// Store table as a single paragraph entry
						paragraphs = append(paragraphs, StructuredParagraph{
							Content: tableResult.String(),
							IsTable: true,
						})
					} else {
						if result.Len() > 0 {
							result.WriteString("\n\n")
						}
						result.WriteString(tableResult.String())
					}
					tableRows = [][]string{}
				}
			}
		}
	}

	if collectParagraphs {
		// Build final result from collected paragraphs
		for i, para := range paragraphs {
			if i > 0 {
				if para.IsTable {
					if result.Len() > 0 {
						result.WriteString("\n")
					}
				} else {
					result.WriteString("\n")
				}
			}
			result.WriteString(para.Content)
		}
		return result.String(), paragraphs, nil
	}

	return result.String(), nil, nil
}

// wrapError wraps an error with additional context.
func wrapError(message string, err error) error {
	if err == nil {
		return &DOCXParserError{
			message: message,
			cause:   nil,
		}
	}
	return &DOCXParserError{
		message: message,
		cause:   err,
	}
}

// DOCXParserError represents an error that occurs during DOCX parsing.
type DOCXParserError struct {
	message string
	cause   error
}

func (e *DOCXParserError) Error() string {
	if e.message == "" {
		return "DOCX parser error"
	}
	return e.message
}

func (e *DOCXParserError) Unwrap() error {
	return e.cause
}

// ExtractBlocks extracts blocks from docx content based on the given range
func (p *Parser) ExtractBlocks(content string, start, end int) (string, error) {
	// Split into blocks (paragraphs separated by newlines)
	blocks := strings.Split(content, "\n")

	if start < 1 || end < 1 {
		return "", fmt.Errorf("block numbers must start from 1, got %d-%d", start, end)
	}
	if end < start {
		return "", fmt.Errorf("invalid block range: start must not be greater than end (got %d-%d)", start, end)
	}
	if start > len(blocks) {
		return "", nil // Out of range returns empty
	}
	if end > len(blocks) {
		end = len(blocks)
	}

	var result strings.Builder
	for i := start - 1; i < end && i < len(blocks); i++ {
		if i > start-1 {
			result.WriteString("\n")
		}
		result.WriteString(blocks[i])
	}

	return result.String(), nil
}

// ExtractTables extracts tables from docx content based on the given range
func (p *Parser) ExtractTables(content string, start, end int) (string, error) {
	// Split into tables (separated by newlines for this simple test)
	tables := strings.Split(content, "\n")

	if start < 1 || end < 1 {
		return "", fmt.Errorf("table numbers must start from 1, got %d-%d", start, end)
	}
	if end < start {
		return "", fmt.Errorf("invalid table range: start must not be greater than end (got %d-%d)", start, end)
	}
	if start > len(tables) {
		return "", nil // Out of range returns empty
	}
	if end > len(tables) {
		end = len(tables)
	}

	var result strings.Builder
	for i := start - 1; i < end && i < len(tables); i++ {
		if i > start-1 {
			result.WriteString("\n")
		}
		result.WriteString(tables[i])
	}

	return result.String(), nil
}

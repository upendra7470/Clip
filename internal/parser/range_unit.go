package parser

// RangeUnit represents the unit of a range (e.g., pages, lines, characters)
type RangeUnit string

const (
	Pages      RangeUnit = "pages"
	Slides     RangeUnit = "slides"
	Paragraphs RangeUnit = "paragraphs"
	Lines      RangeUnit = "lines"
	Rows       RangeUnit = "rows"
	Blocks     RangeUnit = "blocks"
)

package prettytable

import (
	"fmt"
	"strings"
)

// Alignment type for column alignment
// Left, Center, Right
// (not used in this minimal version)
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// Table represents a table with field names and rows
// Only ASCII rendering is implemented for now
type Table struct {
	fieldNames []string
	rows       [][]any
}

// NewTable creates a new empty table
func NewTable() *Table {
	return &Table{}
}

// NewTableWithFields creates a new table with field names
func NewTableWithFields(fields []string) *Table {
	return &Table{fieldNames: fields}
}

// SetFieldNames sets the field (column) names
func (t *Table) SetFieldNames(fields []string) {
	t.fieldNames = fields
}

// FieldNames returns the field names
func (t *Table) FieldNames() []string {
	return t.fieldNames
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []any) error {
	if len(t.fieldNames) > 0 && len(row) != len(t.fieldNames) {
		return fmt.Errorf("row has %d columns, expected %d", len(row), len(t.fieldNames))
	}
	t.rows = append(t.rows, row)
	return nil
}

// String renders the table as ASCII (implements fmt.Stringer)
func (t *Table) String() string {
	return t.RenderASCII()
}

// RenderASCII renders the table as an ASCII string
func (t *Table) RenderASCII() string {
	if len(t.fieldNames) == 0 {
		return "(no fields)"
	}
	// Compute column widths
	colWidths := make([]int, len(t.fieldNames))
	for i, name := range t.fieldNames {
		colWidths[i] = len(name)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			cellStr := fmt.Sprintf("%v", cell)
			if len(cellStr) > colWidths[i] {
				colWidths[i] = len(cellStr)
			}
		}
	}
	// Helper to build a line
	line := func(sep, fill string) string {
		var b strings.Builder
		b.WriteString(sep)
		for i, w := range colWidths {
			b.WriteString(strings.Repeat(fill, w+2))
			b.WriteString(sep)
			if i == len(colWidths)-1 {
				break
			}
		}
		return b.String()
	}
	// Build table
	var b strings.Builder
	b.WriteString(line("+", "-"))
	b.WriteString("\n")
	// Header
	b.WriteString("|")
	for i, name := range t.fieldNames {
		b.WriteString(" ")
		b.WriteString(padString(name, colWidths[i]))
		b.WriteString(" |")
		if i == len(t.fieldNames)-1 {
			break
		}
	}
	b.WriteString("\n")
	b.WriteString(line("+", "-"))
	b.WriteString("\n")
	// Rows
	for _, row := range t.rows {
		b.WriteString("|")
		for i, cell := range row {
			cellStr := fmt.Sprintf("%v", cell)
			b.WriteString(" ")
			b.WriteString(padString(cellStr, colWidths[i]))
			b.WriteString(" |")
			if i == len(row)-1 {
				break
			}
		}
		b.WriteString("\n")
	}
	b.WriteString(line("+", "-"))
	return b.String()
}

// padString pads s with spaces to width w (left aligned)
func padString(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

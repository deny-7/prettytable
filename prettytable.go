package prettytable

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
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

// AddColumn adds a column to the table with the given field name and column data.
func (t *Table) AddColumn(field string, column []any) error {
	if len(t.rows) > 0 && len(column) != len(t.rows) {
		return fmt.Errorf("column has %d rows, expected %d", len(column), len(t.rows))
	}
	// If no field names yet, just add
	t.fieldNames = append(t.fieldNames, field)
	if len(t.rows) == 0 {
		// No rows yet, create them
		for _, val := range column {
			t.rows = append(t.rows, []any{val})
		}
	} else {
		// Add to existing rows
		for i, val := range column {
			t.rows[i] = append(t.rows[i], val)
		}
	}
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

// FromCSV reads CSV data from an io.Reader and returns a new Table.
func FromCSV(r io.Reader, delim rune) (*Table, error) {
	if delim == 0 {
		// Autodetect delimiter from the first line
		buf := make([]byte, 4096)
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		data := string(buf[:n])
		// Try common delimiters
		candidates := []rune{',', ';', '\t', '|'}
		maxCount := 0
		best := ';'
		for _, d := range candidates {
			count := strings.Count(data, string(d))
			if count > maxCount {
				maxCount = count
				best = d
			}
		}
		delim = best
		// Reset reader to include the bytes we just read
		r = io.MultiReader(strings.NewReader(data), r)
	}
	reader := csv.NewReader(r)
	reader.Comma = delim
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("CSV is empty")
	}
	table := NewTableWithFields(records[0])
	for _, row := range records[1:] {
		rowAny := make([]any, len(row))
		for i, v := range row {
			rowAny[i] = v
		}
		table.AddRow(rowAny)
	}
	return table, nil
}

// FromDBRows creates a Table from a *sql.Rows result set.
func FromDBRows(rows *sql.Rows) (*Table, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	table := NewTableWithFields(columns)
	for rows.Next() {
		values := make([]any, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		rowCopy := make([]any, len(values))
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				rowCopy[i] = string(b)
			} else {
				rowCopy[i] = v
			}
		}
		table.AddRow(rowCopy)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return table, nil
}

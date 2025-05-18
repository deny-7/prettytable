package prettytable

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Alignment type for column alignment
// Left, Center, Right
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
	// alignments stores per-column alignment
	alignments map[string]Alignment
	// sortBy and reverseSort for sorting
	sortBy      string
	reverseSort bool
	// rowFilter for filtering
	rowFilter func([]any) bool
	// style holds table style options
	style TableStyle
}

// TableStyle holds options for customizing table appearance
// All fields are optional; zero values mean default behavior
type TableStyle struct {
	Border                  bool
	PreserveInternalBorder  bool
	Header                  bool
	HRule                   string // "FRAME", "HEADER", "ALL", "NONE"
	VRule                   string // "FRAME", "ALL", "NONE"
	IntFormat               string // e.g. ",d" or "03d"
	FloatFormat             string // e.g. ".2f"
	CustomFormat            map[string]func(field string, value any) string
	PaddingWidth            int
	LeftPaddingWidth        int
	RightPaddingWidth       int
	VerticalChar            string
	HorizontalChar          string
	HorizontalAlignChar     string
	JunctionChar            string
	TopJunctionChar         string
	BottomJunctionChar      string
	RightJunctionChar       string
	LeftJunctionChar        string
	TopRightJunctionChar    string
	TopLeftJunctionChar     string
	BottomRightJunctionChar string
	BottomLeftJunctionChar  string
	MinTableWidth           int
	MaxTableWidth           int
	MaxWidth                int
	MinWidth                int
	UseHeaderWidth          *bool
	BreakOnHyphens          *bool
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

// DelRow deletes a row at the given index.
func (t *Table) DelRow(index int) error {
	if index < 0 || index >= len(t.rows) {
		return fmt.Errorf("row index %d out of range", index)
	}
	t.rows = append(t.rows[:index], t.rows[index+1:]...)
	return nil
}

// DelColumn deletes a column by field name.
func (t *Table) DelColumn(field string) error {
	idx := -1
	for i, name := range t.fieldNames {
		if name == field {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("column %q not found", field)
	}
	t.fieldNames = append(t.fieldNames[:idx], t.fieldNames[idx+1:]...)
	for i := range t.rows {
		if idx < len(t.rows[i]) {
			t.rows[i] = append(t.rows[i][:idx], t.rows[i][idx+1:]...)
		}
	}
	return nil
}

// ClearRows deletes all rows but keeps field names.
func (t *Table) ClearRows() {
	t.rows = nil
}

// Clear deletes all rows and field names.
func (t *Table) Clear() {
	t.rows = nil
	t.fieldNames = nil
}

// String renders the table as ASCII (implements fmt.Stringer)
func (t *Table) String() string {
	return t.RenderASCII()
}

// SetAlign sets the alignment for a column by field name.
func (t *Table) SetAlign(field string, align Alignment) {
	if t.alignments == nil {
		t.alignments = make(map[string]Alignment)
	}
	t.alignments[field] = align
}

// SetAlignAll sets the alignment for all columns.
func (t *Table) SetAlignAll(align Alignment) {
	if t.alignments == nil {
		t.alignments = make(map[string]Alignment)
	}
	for _, f := range t.fieldNames {
		t.alignments[f] = align
	}
}

// SetSortBy sets the field to sort by and order.
func (t *Table) SetSortBy(field string, reverse bool) {
	t.sortBy = field
	t.reverseSort = reverse
}

// SetRowFilter sets a filter function for rows.
func (t *Table) SetRowFilter(filter func([]any) bool) {
	t.rowFilter = filter
}

// SetStyle sets the table style options
func (t *Table) SetStyle(style TableStyle) {
	t.style = style
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
	rows := t.rows
	// Filtering
	if t.rowFilter != nil {
		var filtered [][]any
		for _, row := range rows {
			if t.rowFilter(row) {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	// Sorting
	if t.sortBy != "" {
		idx := -1
		for i, name := range t.fieldNames {
			if name == t.sortBy {
				idx = i
				break
			}
		}
		if idx != -1 {
			sorted := make([][]any, len(rows))
			copy(sorted, rows)
			less := func(i, j int) bool {
				si := fmt.Sprintf("%v", sorted[i][idx])
				sj := fmt.Sprintf("%v", sorted[j][idx])
				if t.reverseSort {
					return sj < si
				}
				return si < sj
			}
			sort.Slice(sorted, less)
			rows = sorted
		}
	}
	for i, name := range t.fieldNames {
		colWidths[i] = len(name)
	}
	for _, row := range rows {
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
		align := AlignLeft
		if t.alignments != nil {
			if a, ok := t.alignments[name]; ok {
				align = a
			}
		}
		b.WriteString(" ")
		b.WriteString(padAlign(name, colWidths[i], align))
		b.WriteString(" |")
		if i == len(t.fieldNames)-1 {
			break
		}
	}
	b.WriteString("\n")
	b.WriteString(line("+", "-"))
	b.WriteString("\n")
	// Rows
	for _, row := range rows {
		b.WriteString("|")
		for i, cell := range row {
			cellStr := fmt.Sprintf("%v", cell)
			align := AlignLeft
			if t.alignments != nil {
				if a, ok := t.alignments[t.fieldNames[i]]; ok {
					align = a
				}
			}
			b.WriteString(" ")
			b.WriteString(padAlign(cellStr, colWidths[i], align))
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

// padAlign pads s to width w with the given alignment
func padAlign(s string, w int, align Alignment) string {
	pad := w - len(s)
	if pad <= 0 {
		return s
	}
	switch align {
	case AlignRight:
		return strings.Repeat(" ", pad) + s
	case AlignCenter:
		left := pad / 2
		right := pad - left
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	default:
		return s + strings.Repeat(" ", pad)
	}
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

// RenderText renders the table as plain text (same as ASCII)
func (t *Table) RenderText() string {
	return t.RenderASCII()
}

// RenderCSV renders the table as CSV
func (t *Table) RenderCSV() string {
	var b strings.Builder
	w := csv.NewWriter(&b)
	w.Write(t.fieldNames)
	for _, row := range t.rows {
		rec := make([]string, len(row))
		for i, v := range row {
			rec[i] = fmt.Sprintf("%v", v)
		}
		w.Write(rec)
	}
	w.Flush()
	return b.String()
}

// RenderJSON renders the table as JSON array of objects
func (t *Table) RenderJSON() string {
	objs := make([]map[string]any, len(t.rows))
	for i, row := range t.rows {
		obj := make(map[string]any)
		for j, name := range t.fieldNames {
			if j < len(row) {
				obj[name] = row[j]
			}
		}
		objs[i] = obj
	}
	data, err := json.MarshalIndent(objs, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// RenderHTML renders the table as an HTML table
func (t *Table) RenderHTML() string {
	escape := func(s string) string {
		s = strings.ReplaceAll(s, "&", "&amp;")
		s = strings.ReplaceAll(s, "<", "&lt;")
		s = strings.ReplaceAll(s, ">", "&gt;")
		s = strings.ReplaceAll(s, "\"", "&quot;")
		return s
	}
	var b strings.Builder
	b.WriteString("<table border=\"1\">\n<tr>")
	for _, name := range t.fieldNames {
		b.WriteString("<th>")
		b.WriteString(escape(name))
		b.WriteString("</th>")
	}
	b.WriteString("</tr>\n")
	for _, row := range t.rows {
		b.WriteString("<tr>")
		for i, cell := range row {
			b.WriteString("<td>")
			b.WriteString(escape(fmt.Sprintf("%v", cell)))
			b.WriteString("</td>")
			if i == len(row)-1 {
				break
			}
		}
		b.WriteString("</tr>\n")
	}
	b.WriteString("</table>")
	return b.String()
}

// RenderLaTeX renders the table as LaTeX tabular
func (t *Table) RenderLaTeX() string {
	escape := func(s string) string {
		s = strings.ReplaceAll(s, "\\", "\\textbackslash{}")
		s = strings.ReplaceAll(s, "_", "\\_")
		s = strings.ReplaceAll(s, "&", "\\&")
		s = strings.ReplaceAll(s, "%", "\\%")
		s = strings.ReplaceAll(s, "$", "\\$")
		s = strings.ReplaceAll(s, "#", "\\#")
		s = strings.ReplaceAll(s, "{", "\\{")
		s = strings.ReplaceAll(s, "}", "\\}")
		s = strings.ReplaceAll(s, "~", "\\textasciitilde{}")
		s = strings.ReplaceAll(s, "^", "\\textasciicircum{}")
		return s
	}
	var b strings.Builder
	b.WriteString("\\begin{tabular}{|" + strings.Repeat("l|", len(t.fieldNames)) + "}\n\\hline\n")
	for i, name := range t.fieldNames {
		b.WriteString(escape(name))
		if i < len(t.fieldNames)-1 {
			b.WriteString(" & ")
		}
	}
	b.WriteString(" \\ \\hline\n")
	for _, row := range t.rows {
		for i, cell := range row {
			b.WriteString(escape(fmt.Sprintf("%v", cell)))
			if i < len(row)-1 {
				b.WriteString(" & ")
			}
		}
		b.WriteString(" \\ \\hline\n")
	}
	b.WriteString("\\end{tabular}")
	return b.String()
}

// RenderMediaWiki renders the table as MediaWiki markup
func (t *Table) RenderMediaWiki() string {
	var b strings.Builder
	b.WriteString("{| class=\"wikitable\"\n|-")
	for _, name := range t.fieldNames {
		b.WriteString("! ")
		b.WriteString(name)
		b.WriteString(" ")
	}
	b.WriteString("\n")
	for _, row := range t.rows {
		b.WriteString("|-")
		for _, cell := range row {
			b.WriteString("| ")
			b.WriteString(fmt.Sprintf("%v", cell))
			b.WriteString(" ")
		}
		b.WriteString("\n")
	}
	b.WriteString("|}")
	return b.String()
}

// RenderUnicode renders the table using Unicode box-drawing characters
func (t *Table) RenderUnicode() string {
	if len(t.fieldNames) == 0 {
		return "(no fields)"
	}
	// Compute column widths
	colWidths := make([]int, len(t.fieldNames))
	for i, name := range t.fieldNames {
		colWidths[i] = runeWidth(name)
	}
	rows := t.rows
	// Filtering
	if t.rowFilter != nil {
		var filtered [][]any
		for _, row := range rows {
			if t.rowFilter(row) {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	// Sorting
	if t.sortBy != "" {
		idx := -1
		for i, name := range t.fieldNames {
			if name == t.sortBy {
				idx = i
				break
			}
		}
		if idx != -1 {
			sorted := make([][]any, len(rows))
			copy(sorted, rows)
			less := func(i, j int) bool {
				si := fmt.Sprintf("%v", sorted[i][idx])
				sj := fmt.Sprintf("%v", sorted[j][idx])
				if t.reverseSort {
					return sj < si
				}
				return si < sj
			}
			sort.Slice(sorted, less)
			rows = sorted
		}
	}
	for i, name := range t.fieldNames {
		w := runeWidth(name)
		if w > colWidths[i] {
			colWidths[i] = w
		}
	}
	for _, row := range rows {
		for i, cell := range row {
			cellStr := fmt.Sprintf("%v", cell)
			w := runeWidth(cellStr)
			if w > colWidths[i] {
				colWidths[i] = w
			}
		}
	}
	// Helper to build a line
	line := func(left, mid, right, sep string) string {
		var b strings.Builder
		b.WriteString(left)
		for i, w := range colWidths {
			b.WriteString(strings.Repeat(mid, w+2))
			if i < len(colWidths)-1 {
				b.WriteString(sep)
			}
		}
		b.WriteString(right)
		return b.String()
	}
	// Box-drawing chars

	top := line("┌", "─", "┐", "┬")
	mid := line("├", "─", "┤", "┼")
	bot := line("└", "─", "┘", "┴")
	var b strings.Builder
	b.WriteString(top)
	b.WriteString("\n")
	// Header
	b.WriteString("│")
	for i, name := range t.fieldNames {
		align := AlignLeft
		if t.alignments != nil {
			if a, ok := t.alignments[name]; ok {
				align = a
			}
		}
		b.WriteString(" ")
		b.WriteString(padAlignUnicode(name, colWidths[i], align))
		b.WriteString(" │")
	}
	b.WriteString("\n")
	b.WriteString(mid)
	b.WriteString("\n")
	// Rows
	for _, row := range rows {
		b.WriteString("│")
		for i, cell := range row {
			cellStr := fmt.Sprintf("%v", cell)
			align := AlignLeft
			if t.alignments != nil {
				if a, ok := t.alignments[t.fieldNames[i]]; ok {
					align = a
				}
			}
			b.WriteString(" ")
			b.WriteString(padAlignUnicode(cellStr, colWidths[i], align))
			b.WriteString(" │")
		}
		b.WriteString("\n")
	}
	b.WriteString(bot)
	return b.String()
}

// runeWidth returns the number of runes (Unicode code points) in a string
func runeWidth(s string) int {
	return len([]rune(s))
}

// padAlignUnicode pads s to width w (in runes) with the given alignment
func padAlignUnicode(s string, w int, align Alignment) string {
	r := []rune(s)
	pad := w - len(r)
	if pad <= 0 {
		return s
	}
	switch align {
	case AlignRight:
		return strings.Repeat(" ", pad) + s
	case AlignCenter:
		left := pad / 2
		right := pad - left
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	default:
		return s + strings.Repeat(" ", pad)
	}
}

// htmlEscape escapes HTML special chars
func htmlEscape(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;")
	return replacer.Replace(s)
}

// latexEscape escapes LaTeX special chars
func latexEscape(s string) string {
	replacer := strings.NewReplacer("&", "\\&", "%", "\\%", "$", "\\$", "#", "\\#", "_", "\\_", "{", "\\{", "}", "\\}", "~", "\\textasciitilde{}", "^", "\\textasciicircum{}", "\\", "\\textbackslash{}")
	return replacer.Replace(s)
}

// RenderMarkdown renders the table as GitHub-flavored Markdown
func (t *Table) RenderMarkdown() string {
	if len(t.fieldNames) == 0 {
		return "(no fields)"
	}
	var b strings.Builder
	// Header row
	b.WriteString("| ")
	for i, name := range t.fieldNames {
		b.WriteString(name)
		b.WriteString(" | ")
		if i == len(t.fieldNames)-1 {
			break
		}
	}
	b.WriteString("\n| ")
	// Separator row
	for i := range t.fieldNames {
		b.WriteString("--- | ")
		if i == len(t.fieldNames)-1 {
			break
		}
	}
	b.WriteString("\n")
	// Data rows
	for _, row := range t.rows {
		b.WriteString("| ")
		for i, cell := range row {
			b.WriteString(fmt.Sprintf("%v", cell))
			b.WriteString(" | ")
			if i == len(row)-1 {
				break
			}
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// GetFormattedString returns the table as a string in the specified format.
// Supported formats: "text", "ascii", "csv", "json", "html", "latex", "mediawiki", "markdown"
func (t *Table) GetFormattedString(format string) string {
	switch strings.ToLower(format) {
	case "text", "ascii":
		return t.RenderASCII()
	case "csv":
		return t.RenderCSV()
	case "json":
		return t.RenderJSON()
	case "html":
		return t.RenderHTML()
	case "latex":
		return t.RenderLaTeX()
	case "mediawiki":
		return t.RenderMediaWiki()
	case "markdown":
		return t.RenderMarkdown()
	default:
		return t.RenderASCII()
	}
}

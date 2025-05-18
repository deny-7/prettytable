# prettytable (Go)

A Go port of the Python PrettyTable package. Display tabular data in ASCII, Unicode, Markdown, CSV, HTML, JSON, LaTeX, and MediaWiki formats.

[![Go Reference](https://pkg.go.dev/badge/github.com/deny-7/prettytable.svg)](https://pkg.go.dev/github.com/deny-7/prettytable)

## Features
- Create tables with custom field (column) names
- Add rows one by one or all at once
- Add columns one by one
- Delete rows and columns
- Sort and filter rows
- Control column alignment (left, center, right)
- Section dividers (add_divider)
- Advanced style options (borders, padding, custom chars)
- Output formats: ASCII, Unicode, Markdown, CSV, HTML, JSON, LaTeX, MediaWiki and Markdown
- Import from CSV or database rows

## Installation

```sh
go get github.com/deny-7/prettytable
```

## Usage

### Basic Example

```go
package main

import (
	"fmt"
	"github.com/deny-7/prettytable"
)

func main() {
	t := prettytable.NewTableWithFields([]string{"City name", "Area", "Population", "Annual Rainfall"})
	t.AddRows([][]any{
		{"Adelaide", 1295, 1158259, 600.5},
		{"Brisbane", 5905, 1857594, 1146.4},
		{"Darwin", 112, 120900, 1714.7},
	})
	fmt.Println(t)
}
```

### Adding Data

#### Row by row

```go
t := prettytable.NewTable()
t.SetFieldNames([]string{"City name", "Area", "Population", "Annual Rainfall"})
t.AddRow([]any{"Adelaide", 1295, 1158259, 600.5})
t.AddRow([]any{"Brisbane", 5905, 1857594, 1146.4})
```

#### All rows at once

```go
t := prettytable.NewTableWithFields([]string{"City name", "Area", "Population", "Annual Rainfall"})
t.AddRows([][]any{
	{"Adelaide", 1295, 1158259, 600.5},
	{"Brisbane", 5905, 1857594, 1146.4},
	{"Darwin", 112, 120900, 1714.7},
})
```

#### Column by column

```go
t := prettytable.NewTable()
t.AddColumn("City name", []any{"Adelaide", "Brisbane", "Darwin"})
t.AddColumn("Area", []any{1295, 5905, 112})
t.AddColumn("Population", []any{1158259, 1857594, 120900})
t.AddColumn("Annual Rainfall", []any{600.5, 1146.4, 1714.7})
```

#### Importing from CSV

```go
f, _ := os.Open("myfile.csv")
table, _ := prettytable.FromCSV(f, ',')
```

#### Importing from database rows

```go
rows, _ := db.Query("SELECT name, area, population, rainfall FROM cities")
table, _ := prettytable.FromDBRows(rows)
```

### Output Formats

```go
fmt.Println(table.RenderASCII())      // ASCII
fmt.Println(table.RenderUnicode())    // Unicode box-drawing
fmt.Println(table.RenderMarkdown())   // Markdown
fmt.Println(table.RenderCSV())        // CSV
fmt.Println(table.RenderJSON())       // JSON
fmt.Println(table.RenderHTML())       // HTML
fmt.Println(table.RenderLaTeX())      // LaTeX
fmt.Println(table.RenderMediaWiki())  // MediaWiki
fmt.Println(table.RenderMarkdown())  // Markdown
```

Or use:

```go
fmt.Println(table.GetFormattedString("markdown"))
```

### Advanced Features

#### Section Dividers

```go
t.AddRow([]any{"Adelaide", 1295, 1158259, 600.5})
t.AddDivider()
t.AddRow([]any{"Brisbane", 5905, 1857594, 1146.4})
```

#### Sorting and Filtering

```go
t.SetSortBy("Population", true) // Sort by Population descending
t.SetRowFilter(func(row []any) bool { return row[2].(int) > 1000000 }) // Only large cities
```

#### Alignment

```go
t.SetAlign("City name", prettytable.AlignLeft)
t.SetAlign("Population", prettytable.AlignRight)
t.SetAlignAll(prettytable.AlignCenter)
```

#### Custom Style

```go
style := prettytable.TableStyle{
	Border:         false,
	PaddingWidth:   1,
	VerticalChar:   ".",
	HorizontalChar: "_",
	JunctionChar:   "*",
}
t.SetStyle(style)
```

## Example Output

**ASCII:**
```
+-----------+------+------------+-----------------+
| City name | Area | Population | Annual Rainfall |
+-----------+------+------------+-----------------+
| Adelaide  | 1295 | 1158259    | 600.5           |
| Brisbane  | 5905 | 1857594    | 1146.4          |
| Darwin    | 112  | 120900     | 1714.7          |
+-----------+------+------------+-----------------+
```

**Markdown:**
```
| City name | Area | Population | Annual Rainfall |
| --- | --- | --- | --- |
| Adelaide | 1295 | 1158259 | 600.5 |
| Brisbane | 5905 | 1857594 | 1146.4 |
| Darwin | 112 | 120900 | 1714.7 |
```

**Unicode:**
```
┌───────────┬──────┬────────────┬─────────────────┐
│ City name │ Area │ Population │ Annual Rainfall │
├───────────┼──────┼────────────┼─────────────────┤
│ Adelaide  │ 1295 │ 1158259    │ 600.5           │
│ Brisbane  │ 5905 │ 1857594    │ 1146.4          │
│ Darwin    │ 112  │ 120900     │ 1714.7          │
└───────────┴──────┴────────────┴─────────────────┘
```

## API Reference

See GoDoc: https://pkg.go.dev/github.com/deny-7/prettytable

## License
Ported to Go by Denys Bondar, 2025  
Based on original work by Luke Maurits and contributors  
Licensed under the BSD 3-Clause License  

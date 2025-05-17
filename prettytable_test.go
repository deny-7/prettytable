package prettytable

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestTableBasic(t *testing.T) {
	table := NewTableWithFields([]string{"A", "B"})
	err := table.AddRow([]any{"foo", 123})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = table.AddRow([]any{"bar", 456})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `+-----+-----+
| A   | B   |
+-----+-----+
| foo | 123 |
| bar | 456 |
+-----+-----+`
	actual := strings.TrimSpace(table.RenderASCII())
	if actual != expected {
		t.Errorf("ASCII output mismatch.\nExpected:\n%s\nActual:\n%s", expected, actual)
	}
}

func TestTableFieldNames(t *testing.T) {
	table := NewTable()
	table.SetFieldNames([]string{"X", "Y"})
	if got := table.FieldNames(); len(got) != 2 || got[0] != "X" || got[1] != "Y" {
		t.Errorf("FieldNames() = %v, want [X Y]", got)
	}
}

func TestTableAddRowError(t *testing.T) {
	table := NewTableWithFields([]string{"A", "B"})
	err := table.AddRow([]any{"only one col"})
	if err == nil {
		t.Error("expected error for wrong column count, got nil")
	}
}

func TestTableAddColumn(t *testing.T) {
	table := NewTable()
	err := table.AddColumn("City name", []any{"Adelaide", "Brisbane", "Darwin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = table.AddColumn("Area", []any{1295, 5905, 112})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = table.AddColumn("Population", []any{1158259, 1857594, 120900})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = table.AddColumn("Annual Rainfall", []any{600.5, 1146.4, 1714.7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `+-----------+------+------------+-----------------+
| City name | Area | Population | Annual Rainfall |
+-----------+------+------------+-----------------+
| Adelaide  | 1295 | 1158259    | 600.5           |
| Brisbane  | 5905 | 1857594    | 1146.4          |
| Darwin    | 112  | 120900     | 1714.7          |
+-----------+------+------------+-----------------+`
	actual := strings.TrimSpace(table.RenderASCII())
	if actual != expected {
		t.Errorf("ASCII output mismatch.\nExpected:\n%s\nActual:\n%s", expected, actual)
	}
}

func TestTableAddColumnError(t *testing.T) {
	table := NewTable()
	table.AddColumn("A", []any{1, 2, 3})
	err := table.AddColumn("B", []any{4, 5}) // wrong length
	if err == nil {
		t.Error("expected error for wrong column length, got nil")
	}
}

func TestFromCSV(t *testing.T) {
	csvData := `City name,Area,Population,Annual Rainfall
Adelaide,1295,1158259,600.5
Brisbane,5905,1857594,1146.4
Darwin,112,120900,1714.7`
	r := strings.NewReader(csvData)
	table, err := FromCSV(r, ',')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `+-----------+------+------------+-----------------+
| City name | Area | Population | Annual Rainfall |
+-----------+------+------------+-----------------+
| Adelaide  | 1295 | 1158259    | 600.5           |
| Brisbane  | 5905 | 1857594    | 1146.4          |
| Darwin    | 112  | 120900     | 1714.7          |
+-----------+------+------------+-----------------+`
	actual := strings.TrimSpace(table.RenderASCII())
	if actual != expected {
		t.Errorf("ASCII output mismatch.\nExpected:\n%s\nActual:\n%s", expected, actual)
	}
}

func TestFromCSV_Empty(t *testing.T) {
	r := strings.NewReader("")
	_, err := FromCSV(r, ',')
	if err == nil {
		t.Error("expected error for empty CSV, got nil")
	}
}

func TestFromDBRows_SQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE cities (
		name TEXT, area INTEGER, population INTEGER, rainfall REAL
	)`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = db.Exec(`INSERT INTO cities (name, area, population, rainfall) VALUES
		('Adelaide', 1295, 1158259, 600.5),
		('Brisbane', 5905, 1857594, 1146.4),
		('Darwin', 112, 120900, 1714.7)
	`)
	if err != nil {
		t.Fatalf("failed to insert data: %v", err)
	}

	rows, err := db.Query("SELECT name, area, population, rainfall FROM cities")
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	defer rows.Close()

	table, err := FromDBRows(rows)
	if err != nil {
		t.Fatalf("FromDBRows error: %v", err)
	}

	expected := `+----------+------+------------+----------+
| name     | area | population | rainfall |
+----------+------+------------+----------+
| Adelaide | 1295 | 1158259    | 600.5    |
| Brisbane | 5905 | 1857594    | 1146.4   |
| Darwin   | 112  | 120900     | 1714.7   |
+----------+------+------------+----------+`
	actual := strings.TrimSpace(table.RenderASCII())
	if actual != expected {
		t.Errorf("ASCII output mismatch.\nExpected:\n%s\nActual:\n%s", expected, actual)
	}
}

func TestDelRowAndDelColumn(t *testing.T) {
	table := NewTableWithFields([]string{"A", "B", "C"})
	table.AddRow([]any{1, 2, 3})
	table.AddRow([]any{4, 5, 6})
	table.AddRow([]any{7, 8, 9})

	// Remove middle row
	err := table.DelRow(1)
	if err != nil {
		t.Fatalf("DelRow error: %v", err)
	}
	if len(table.rows) != 2 || table.rows[1][0] != 7 {
		t.Errorf("DelRow did not remove row correctly: %+v", table.rows)
	}

	// Remove first column
	err = table.DelColumn("A")
	if err != nil {
		t.Fatalf("DelColumn error: %v", err)
	}
	if len(table.fieldNames) != 2 || table.fieldNames[0] != "B" {
		t.Errorf("DelColumn did not remove column correctly: %+v", table.fieldNames)
	}
	if table.rows[0][0] != 2 || table.rows[1][0] != 8 {
		t.Errorf("DelColumn did not update rows correctly: %+v", table.rows)
	}

	// Error cases
	if err := table.DelRow(10); err == nil {
		t.Error("expected error for out-of-range row index")
	}
	if err := table.DelColumn("Z"); err == nil {
		t.Error("expected error for missing column name")
	}
}

func TestClearRowsAndClear(t *testing.T) {
	table := NewTableWithFields([]string{"A", "B"})
	table.AddRow([]any{1, 2})
	table.AddRow([]any{3, 4})
	table.ClearRows()
	if len(table.rows) != 0 {
		t.Errorf("ClearRows did not clear rows: %+v", table.rows)
	}
	if len(table.fieldNames) != 2 {
		t.Errorf("ClearRows should not clear field names: %+v", table.fieldNames)
	}
	table.Clear()
	if len(table.rows) != 0 || len(table.fieldNames) != 0 {
		t.Errorf("Clear did not clear table: rows=%+v, fields=%+v", table.rows, table.fieldNames)
	}
}

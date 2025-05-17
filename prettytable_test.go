package prettytable

import (
	"strings"
	"testing"
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

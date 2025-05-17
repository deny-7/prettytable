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

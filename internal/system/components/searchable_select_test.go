package components

import (
	"bytes"
	"strings"
	"testing"
)

func TestSearchableSelectRendersAccessibleSelectedValueAndError(t *testing.T) {
	var output bytes.Buffer
	node := SearchableSelect("owner_id", "owner_id", "Dueño", "Busca un dueño", "42", []SearchableOption{
		{Value: "42", Label: "María Pérez · 123456"},
		{Value: "43", Label: "Juan Gómez · 654321"},
	}, true, "Selecciona un dueño válido.")
	if err := node.Render(&output); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	for _, want := range []string{
		`data-controller="searchable-select"`,
		`<label for="owner_id"`,
		`id="owner_id"`,
		`role="combobox"`,
		`data-action="input-&gt;searchable-select#clearValidity focus-&gt;searchable-select#selectValue"`,
		`aria-controls="owner_id-listbox"`,
		`aria-invalid="true"`,
		`aria-describedby="owner_id-error"`,
		`<input type="hidden" name="owner_id" value="42">`,
		`role="option" data-value="42" aria-selected="true"`,
		`Selecciona un dueño válido.`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("searchable select missing %q in %s", want, html)
		}
	}
}

func TestSearchableSelectKeepsEmptyRequiredValueUnselected(t *testing.T) {
	var output bytes.Buffer
	node := SearchableSelect("cabin_id", "cabin_id", "Cabaña", "Busca una cabaña", "", []SearchableOption{
		{Value: "7", Label: "Refugio del río"},
	}, true, "")
	if err := node.Render(&output); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	if !strings.Contains(html, `placeholder="Busca una cabaña"`) || !strings.Contains(html, `name="cabin_id" value=""`) || !strings.Contains(html, `required aria-required="true"`) {
		t.Fatalf("empty searchable select lost its prompt or submitted value: %s", html)
	}
	if strings.Contains(html, `aria-selected="true"`) {
		t.Fatalf("empty searchable select marked an option selected: %s", html)
	}
}

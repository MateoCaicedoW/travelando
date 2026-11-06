package components

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestAmenityDeleteUsesStimulusConfirmation(t *testing.T) {
	var html strings.Builder
	if err := AmenityManager("Amenidades", "Descripción", "/admin/amenities", []models.AmenityType{{ID: 1, Name: "Piscina"}}, nil).Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), `data-confirmation-message="¿Eliminar amenidad y sus valores de todas las cabañas?"`) || strings.Contains(html.String(), "onsubmit=") {
		t.Fatalf("amenity delete should use shared Stimulus confirmation: %s", html.String())
	}
}

func TestAmenityEditorHasMobileCardLabels(t *testing.T) {
	var html strings.Builder
	if err := AmenityManager("Amenidades", "Descripción", "/admin/amenities", []models.AmenityType{{ID: 1, Name: "Piscina"}}, nil).Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"responsive-data-table", `data-label="Amenidad"`, `data-label="Tipo de valor"`, `data-label="Acciones"`, "md:flex-nowrap", "md:min-h-9", "w-48", "shrink-0", "Piscina", "Guardar", "Eliminar"} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("amenity editor missing %q", want)
		}
	}
}

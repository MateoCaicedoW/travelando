package dashboard

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

func TestInventoryEditOpensCabinModal(t *testing.T) {
	var html strings.Builder
	page := services.CabinPageResult{Cabins: []models.Cabin{{ID: 7, Name: "Cabaña"}}}
	if err := inventoryEl(page).Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`hx-get="/admin/cabins/7/edit"`, `hx-target="body"`, `hx-swap="beforeend"`} {
		if !strings.Contains(html.String(), expected) {
			t.Errorf("inventory edit link missing %q", expected)
		}
	}
}

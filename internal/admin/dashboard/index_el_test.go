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

func TestInventoryShowsOnlyAcceptedCabinPrice(t *testing.T) {
	var html strings.Builder
	page := services.CabinPageResult{Cabins: []models.Cabin{{
		ID: 7, Name: "Cabaña", AcceptsDayUse: true, DayPrice: 100000, NightPrice: 200000,
	}}}
	if err := inventoryEl(page).Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "Pasadía $100.000") || strings.Contains(html.String(), "Noche $200.000") {
		t.Fatalf("inventory showed prices for an unaccepted stay type: %s", html.String())
	}
}

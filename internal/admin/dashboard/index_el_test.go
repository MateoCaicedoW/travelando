package dashboard

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

func TestDashboardActivityHasMobileCardLabels(t *testing.T) {
	var html strings.Builder
	page := dashboardEl([]models.Inquiry{{
		ID: 9, Status: models.InquiryReserved, GuestName: "Huésped", GuestPhone: "573001234567", CabinName: "Cabaña", OwnerName: "Dueño",
		DateFrom: "2099-08-10", DateTo: "2099-08-11", Commission: 12500,
	}}, 0, 0, nil, nil, 1)
	if err := page.Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"responsive-data-table", `data-label="Huésped"`, `data-label="Cabaña / contacto"`, `data-label="Fechas"`, `data-label="Estado"`, `data-label="Comisión"`, "12.500"} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("dashboard activity missing %q", want)
		}
	}
}

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

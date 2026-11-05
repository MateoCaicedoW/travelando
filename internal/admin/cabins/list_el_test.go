package cabins

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

func TestAdminCabinCardUsesSharedPhotoGallery(t *testing.T) {
	page := services.CabinPageResult{Total: 1, Cabins: []models.Cabin{{
		ID: 7, Name: "Cabaña", Location: "Montaña", OwnerName: "Dueña", OwnerWhatsApp: "573001234567",
		OwnerID: 1, OwnerActive: true, Active: true, IncludedGuests: 2,
		Photos: []models.CabinPhoto{{URL: "/fotos/1"}, {URL: "/fotos/2"}},
	}}}
	var html strings.Builder
	if err := cabinListResultsEl(page, "").Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"/fotos/1", "/fotos/2", "Fotos de Cabaña", "Dueña · WhatsApp 573001234567", "Editar cabaña"} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("admin cabin card missing %q: %s", want, html.String())
		}
	}
}

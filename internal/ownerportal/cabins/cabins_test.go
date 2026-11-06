package cabins

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestOwnerCabinCardUsesSharedPhotoGallery(t *testing.T) {
	cabin := models.Cabin{
		ID: 7, Name: "Cabaña", Location: "Montaña", Description: "Vista al valle", Active: true,
		IncludedGuests: 2, DayPrice: 100000, NightPrice: 200000,
		Photos: []models.CabinPhoto{{URL: "/fotos/1"}, {URL: "/fotos/2"}},
	}
	var html strings.Builder
	if err := ownerCabinCardEl(cabin, 1).Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"/fotos/1", "/fotos/2", "Fotos de Cabaña", "Vista al valle", "Editar cabaña"} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("owner cabin card missing %q: %s", want, html.String())
		}
	}
}

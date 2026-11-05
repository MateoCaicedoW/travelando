package reservations

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestReservationDetailsShowGuestCedula(t *testing.T) {
	var html strings.Builder
	inquiry := models.Inquiry{GuestName: "Huésped", GuestCedula: "12345678", TermsAcceptedAt: "2026-11-05 12:00:00", Status: models.InquiryPending}
	if err := OwnerReservationDetailsEl(inquiry).Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "Cédula") || !strings.Contains(html.String(), "12345678") || !strings.Contains(html.String(), "Términos aceptados") || !strings.Contains(html.String(), "2026-11-05 12:00:00") {
		t.Fatalf("owner details should show guest identity and terms acceptance: %s", html.String())
	}
}

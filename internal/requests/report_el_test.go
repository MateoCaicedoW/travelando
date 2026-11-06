package requests

import (
	"strings"
	"testing"

	"go.leapkit.dev/core/form/validate"
	"travelando/internal/system/models"
)

func TestReportShowsCancelledReservationStatus(t *testing.T) {
	var html strings.Builder
	inquiry := models.Inquiry{Status: models.InquiryCancelled, CancelledFromStatus: models.InquiryPending, StayType: models.StayDayUse, DateFrom: "2099-02-10", DateTo: "2099-02-11"}
	if err := reportPage(inquiry, models.Cabin{}, "report-token", nil, validate.Errors{}).Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "Solicitud cancelada") || strings.Contains(html.String(), "Pendiente de confirmar") {
		t.Fatalf("report status = %s, want cancelled", html.String())
	}
}

func TestReportShowsCabinDetailsWithoutCommission(t *testing.T) {
	var html strings.Builder
	inquiry := models.Inquiry{Status: models.InquiryReserved, Commission: 98765}
	cabin := models.Cabin{
		Name: "Refugio del río", Location: "Guatapé", Description: "Vista al agua",
		IncludedGuests: 4, AcceptsDayUse: true, DayStart: "09:00", DayEnd: "18:00", DayPrice: 120000,
		Photos:    []models.CabinPhoto{{URL: "/cabin.jpg"}},
		Amenities: []models.CabinAmenity{{Name: "Piscina", ValueType: models.AmenityBoolean, BooleanValue: true}},
	}
	if err := reportPage(inquiry, cabin, "report-token", nil, validate.Errors{}).Render(&html); err != nil {
		t.Fatal(err)
	}
	got := html.String()
	for _, want := range []string{"Refugio del río", "Guatapé", "Vista al agua", "Piscina", "4 huéspedes incluidos", "09:00 a 18:00", "/cabin.jpg"} {
		if !strings.Contains(got, want) {
			t.Errorf("report missing cabin detail %q: %s", want, got)
		}
	}
	for _, want := range []string{"lg:h-full", "justify-center", "rounded-full bg-cream px-3 py-1.5"} {
		if !strings.Contains(got, want) {
			t.Errorf("report missing cabin card styling %q: %s", want, got)
		}
	}
	for _, unexpected := range []string{"Comisión", "98.765", "120.000"} {
		if strings.Contains(got, unexpected) {
			t.Errorf("report exposes sensitive or duplicate price %q: %s", unexpected, got)
		}
	}
}

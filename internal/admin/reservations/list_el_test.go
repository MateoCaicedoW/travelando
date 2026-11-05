package reservations

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestReservationsListIncludesCancelledHistoryAndDetails(t *testing.T) {
	inquiries := reservationEntries([]models.Inquiry{
		{ID: 1, Status: models.InquiryLost},
		{ID: 2, Status: models.InquiryLost},
		{ID: 3, Status: models.InquiryPending, Estimated: 125000, GuestName: "Solicitud activa", CabinName: "Cabaña"},
		{ID: 4, Status: models.InquiryReserved, GuestName: "Huésped activo", CabinName: "Cabaña"},
		{ID: 5, Status: models.InquiryCancelled, GuestName: "Huésped anterior", CabinName: "Otra cabaña"},
	})
	if len(inquiries) != 3 {
		t.Fatalf("listed reservation count = %d, want pending, reserved, and cancelled", len(inquiries))
	}
	var html strings.Builder
	if err := reservationsContentEl(inquiries).Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Solicitud activa", "Huésped activo", "Huésped anterior", "Pendiente", "Reservada", "Cancelada", "Ver", "Editar", "Cancelar", `href="/admin/reservations/3"`, `href="/admin/reservations/4"`, `href="/admin/reservations/5"`, `action="/admin/reservations/3/cancel"`, `href="/admin/inquiries/3/edit?return_to=reservations"`, "125.000"} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("reservations list missing %q", want)
		}
	}
}

func TestCancelledReservationDetailsHaveNoEditOrCancelActions(t *testing.T) {
	var html strings.Builder
	if err := reservationDetailsEl(models.Inquiry{
		ID: 8, Status: models.InquiryCancelled, GuestName: "Huésped", StayType: models.StayDayUse,
		DateFrom: "2099-02-10", DateTo: "2099-02-11",
		CommissionAdjustments: []models.CommissionAdjustment{{PreviousAmount: 10000, NewAmount: 0, Amount: -10000, ActorRole: "admin"}},
	}).Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "Reserva cancelada") || !strings.Contains(html.String(), "10/02/2099") || strings.Contains(html.String(), "11/02/2099") || !strings.Contains(html.String(), "10.000 → $0") || strings.Contains(html.String(), "Editar reserva") || strings.Contains(html.String(), "Cancelar reserva") {
		t.Fatalf("cancelled reservation details expose invalid actions: %s", html.String())
	}
}

func TestConfirmedReservationDetailsOfferEditAndCancel(t *testing.T) {
	var html strings.Builder
	if err := reservationDetailsEl(models.Inquiry{ID: 42, Status: models.InquiryReserved}).Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`href="/admin/inquiries/42/edit?return_to=reservations"`, `action="/admin/reservations/42/cancel"`, "Cancelar reserva"} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("confirmed reservation details missing %q", want)
		}
	}
}

func TestPendingInquiryDetailsOfferEditAndCancel(t *testing.T) {
	var html strings.Builder
	if err := reservationDetailsEl(models.Inquiry{ID: 17, Status: models.InquiryPending}).Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Solicitud pendiente", "Editar solicitud", `action="/admin/reservations/17/cancel"`, "Cancelar solicitud"} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("pending inquiry details missing %q", want)
		}
	}
}

func TestReservationDetailsSeparateStatusFromDetailSections(t *testing.T) {
	var html strings.Builder
	if err := reservationDetailsEl(models.Inquiry{ID: 42, Status: models.InquiryReserved}).Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), `class="mb-6 flex items-start justify-between gap-4"`) {
		t.Fatal("reservation details header does not leave space before the detail sections")
	}
}

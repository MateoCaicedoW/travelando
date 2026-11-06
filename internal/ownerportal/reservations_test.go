package ownerportal

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestOwnerReservationsPageShowsOwnedBookingAndEditAction(t *testing.T) {
	var html strings.Builder
	page := ownerReservationsPageEl([]models.Inquiry{
		{ID: 42, Status: models.InquiryReserved, CabinName: "Cabaña del bosque", GuestName: "Huésped", GuestPhone: "573001234567", StayType: models.StayDayUse, DateFrom: "2099-08-10", DateTo: "2099-08-11", Guests: 3, FinalTotal: 125000},
		{ID: 43, Status: models.InquiryPending, CabinName: "Cabaña del bosque", GuestName: "Solicitud", GuestPhone: "573001111111", StayType: models.StayDayUse, DateFrom: "2099-08-11", DateTo: "2099-08-12", Guests: 2, Estimated: 100000},
		{ID: 44, Status: models.InquiryCancelled, CancelledFromStatus: models.InquiryReserved, CabinName: "Cabaña del bosque", GuestName: "Cancelada", GuestPhone: "573001222222", StayType: models.StayDayUse, DateFrom: "2099-08-12", DateTo: "2099-08-13", Guests: 2, FinalTotal: 100000},
	})
	if err := page.Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Reservas de mis cabañas", "responsive-data-table", `data-label="Huésped"`, `data-label="Cabaña"`, `data-label="Fechas y estadía"`, `data-label="Total / estimado"`, `data-label="Estado"`, `data-label="Acciones"`, `class="dropdown-menu relative inline-flex"`, `aria-haspopup="menu"`, `role="menu"`, `role="menuitem"`, "Huésped", "Solicitud", "Cancelada", "Cabaña del bosque", "10/08/2099", "125.000", "Pendiente", "Estimado", `href="/owner/reservations/42/edit?return_to=reservations"`, `href="/owner/reservations/43/edit?return_to=reservations"`, `href="/owner/reservations/43/confirm"`, `href="/owner/reservations/44"`, "/owner/reservations/42/cancel", "Confirmar reserva", "Ver", "Editar", "Cancelar"} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("owner reservations page missing %q", want)
		}
	}
	if strings.Contains(html.String(), "Ver reporte") || strings.Contains(html.String(), "/report") || strings.Contains(html.String(), `href="/owner/reservations/42/confirm"`) || strings.Contains(html.String(), `href="/owner/reservations/44/confirm"`) {
		t.Fatalf("report action or confirmation on non-pending reservation found: %s", html.String())
	}
}

func TestCancelledOwnerReservationOnlyOffersView(t *testing.T) {
	var html strings.Builder
	if err := ownerReservationRow(models.Inquiry{ID: 44, Status: models.InquiryCancelled}).Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), `href="/owner/reservations/44"`) || strings.Contains(html.String(), "/report") || strings.Contains(html.String(), "/confirm") || strings.Contains(html.String(), "/edit") || strings.Contains(html.String(), "/cancel") {
		t.Fatalf("cancelled owner row exposes edit/cancel actions: %s", html.String())
	}
}

func TestOwnerReservationsPageHasEmptyState(t *testing.T) {
	var html strings.Builder
	if err := ownerReservationsPageEl(nil).Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "Aún no tienes reservas") || !strings.Contains(html.String(), "solicitudes pendientes") {
		t.Fatalf("owner reservations page missing empty state: %s", html.String())
	}
}

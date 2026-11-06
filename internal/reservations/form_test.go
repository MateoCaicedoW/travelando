package reservations

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestUpdateFromFormParsesManualTotalAndDayUse(t *testing.T) {
	values := url.Values{
		"guest_name": {"Huésped"}, "guest_phone": {"+57 300 123 4567"}, "guest_email": {"guest@example.com"},
		"stay_type": {string(models.StayDayUse)}, "date_from": {"2027-02-10"}, "date_to": {"2027-02-10"},
		"guests": {"5"}, "final_total": {"150000"}, "use_final_total": {"on"},
	}
	r := httptest.NewRequest("POST", "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatal(err)
	}
	update, fieldErrors := updateFromForm(r)
	if fieldErrors.HasAny() {
		t.Fatalf("unexpected form errors: %v", fieldErrors)
	}
	if update.Guests != 5 || update.FinalTotal != 150000 || !update.UseFinalTotal || update.DateTo != update.DateFrom {
		t.Fatalf("parsed update = %+v, want five guests, manual total, and same-day pasadía", update)
	}
}

func TestPendingUpdateFromFormDoesNotRequireFinalTotal(t *testing.T) {
	values := url.Values{
		"guest_name": {"Huésped"}, "guest_phone": {"573001234567"}, "guest_email": {"guest@example.com"},
		"stay_type": {string(models.StayOvernight)}, "date_from": {"2027-02-10"}, "date_to": {"2027-02-12"}, "guests": {"3"},
	}
	r := httptest.NewRequest("POST", "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatal(err)
	}
	update, fieldErrors := pendingUpdateFromForm(r)
	if fieldErrors.HasAny() {
		t.Fatalf("pending update form errors: %v", fieldErrors)
	}
	if update.GuestName != "Huésped" || update.StayType != models.StayOvernight || update.DateFrom != "2027-02-10" || update.DateTo != "2027-02-12" || update.Guests != 3 {
		t.Fatalf("pending update = %+v, want all request fields", update)
	}
}

func TestReservationFormShowsGuestContactAndCommissionHistory(t *testing.T) {
	inquiry := models.Inquiry{
		ID: 42, GuestName: "Huésped", GuestPhone: "573001234567", GuestEmail: "guest@example.com",
		StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-11", Guests: 2,
		FinalTotal: 100000, CommissionBP: 1000, Commission: 10000,
		CommissionAdjustments: []models.CommissionAdjustment{{PreviousAmount: 8000, NewAmount: 10000, Amount: 2000, ActorRole: "owner", CreatedAt: "2026-11-05 12:00:00"}},
	}
	var html strings.Builder
	if err := reservationForm(inquiry, nil, true, "").Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"guest@example.com", "573001234567", "Ajustes de comisión", "Dueño", "/owner/reservations/42", "modal#close", "<svg"} {
		if !strings.Contains(html.String(), expected) {
			t.Errorf("reservation form missing %q", expected)
		}
	}
	if strings.Contains(html.String(), "×") {
		t.Error("reservation form close button should use a Lucide icon instead of a text glyph")
	}
}

func TestPendingInquiryFormRecalculatesEstimateInsteadOfEditingFinalTotal(t *testing.T) {
	inquiry := models.Inquiry{ID: 19, Status: models.InquiryPending, Estimated: 125000, StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-11"}
	var html strings.Builder
	if err := reservationForm(inquiry, nil, false, "reservations").Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "Editar solicitud") || !strings.Contains(html.String(), "125.000") || !strings.Contains(html.String(), "recalcularemos el estimado") || strings.Contains(html.String(), `name="final_total"`) {
		t.Fatalf("pending inquiry form does not show recalculated estimate workflow: %s", html.String())
	}
}

func TestReservationEditAlwaysRendersModal(t *testing.T) {
	inquiry := models.Inquiry{ID: 7, StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-11"}
	for _, htmx := range []bool{false, true} {
		r := httptest.NewRequest("GET", "/?return_to=reservations", nil)
		if htmx {
			r.Header.Set("HX-Request", "true")
		}
		response := httptest.NewRecorder()
		render(response, r, inquiry, 0, nil, 200)
		if !strings.Contains(response.Body.String(), `id="modal"`) || strings.Contains(response.Body.String(), "<!DOCTYPE html>") {
			t.Fatalf("htmx=%t rendered non-modal edit response: %s", htmx, response.Body.String())
		}
		if !strings.Contains(response.Body.String(), `name="return_to" value="reservations"`) {
			t.Fatalf("htmx=%t edit form did not preserve its return destination", htmx)
		}
	}
}

func TestReservationEditFormAvoidsNestedCardPadding(t *testing.T) {
	var html strings.Builder
	inquiry := models.Inquiry{ID: 7, Status: models.InquiryPending, StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-11"}
	if err := reservationForm(inquiry, nil, false, "reservations").Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), `class="w-full px-4 py-5 sm:px-6 sm:py-6"`) || strings.Contains(html.String(), `rounded-2xl border border-line bg-white p-4 sm:p-6`) {
		t.Fatalf("reservation edit form should use one compact content padding, got: %s", html.String())
	}
}

func TestOwnerReservationEditPreservesReservationsListDestination(t *testing.T) {
	r := httptest.NewRequest("GET", "/?return_to=reservations", nil)
	response := httptest.NewRecorder()
	render(response, r, models.Inquiry{ID: 12, Status: models.InquiryReserved}, 1, nil, 200)
	if !strings.Contains(response.Body.String(), `name="return_to" value="reservations"`) {
		t.Fatalf("owner edit modal does not preserve reservations list destination: %s", response.Body.String())
	}
}

func TestOwnerReservationDetailsUseScopedActionsAndHideActionsWhenCancelled(t *testing.T) {
	var active, cancelled strings.Builder
	if err := OwnerReservationDetailsEl(models.Inquiry{ID: 42, Status: models.InquiryReserved}).Render(&active); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`href="/owner/reservations/42/edit?return_to=reservations"`, `action="/owner/reservations/42/cancel"`, `hx-target="#owner-reservations-content"`} {
		if !strings.Contains(active.String(), want) {
			t.Errorf("owner reservation details missing %q", want)
		}
	}
	if err := OwnerReservationDetailsEl(models.Inquiry{ID: 43, Status: models.InquiryCancelled}).Render(&cancelled); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(cancelled.String(), "/edit") || strings.Contains(cancelled.String(), "/cancel") {
		t.Fatalf("cancelled reservation details expose edit/cancel actions: %s", cancelled.String())
	}
}

func TestUpdateFromFormRejectsInvalidStayRange(t *testing.T) {
	values := url.Values{
		"guest_name": {"Huésped"}, "guest_phone": {"573001234567"}, "guest_email": {"guest@example.com"},
		"stay_type": {string(models.StayDayUse)}, "date_from": {"2027-02-10"}, "date_to": {"2027-02-11"},
		"guests": {"5"}, "final_total": {"150000"},
	}
	r := httptest.NewRequest("POST", "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatal(err)
	}
	_, fieldErrors := updateFromForm(r)
	if !fieldErrors.Has("date_to") {
		t.Fatalf("field errors = %v, want invalid pasadía end date", fieldErrors)
	}
}

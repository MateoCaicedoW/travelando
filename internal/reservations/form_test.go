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

func TestReservationFormShowsGuestContactAndCommissionHistory(t *testing.T) {
	inquiry := models.Inquiry{
		ID: 42, GuestName: "Huésped", GuestPhone: "573001234567", GuestEmail: "guest@example.com",
		StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-11", Guests: 2,
		FinalTotal: 100000, CommissionBP: 1000, Commission: 10000,
		CommissionAdjustments: []models.CommissionAdjustment{{PreviousAmount: 8000, NewAmount: 10000, Amount: 2000, ActorRole: "owner", CreatedAt: "2026-11-05 12:00:00"}},
	}
	var html strings.Builder
	if err := reservationForm(inquiry, nil, true).Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"guest@example.com", "573001234567", "Ajustes de comisión", "Dueño", "/owner/reservations/42", "modal#close"} {
		if !strings.Contains(html.String(), expected) {
			t.Errorf("reservation form missing %q", expected)
		}
	}
}

func TestReservationEditAlwaysRendersModal(t *testing.T) {
	inquiry := models.Inquiry{ID: 7, StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-11"}
	for _, htmx := range []bool{false, true} {
		r := httptest.NewRequest("GET", "/", nil)
		if htmx {
			r.Header.Set("HX-Request", "true")
		}
		response := httptest.NewRecorder()
		render(response, r, inquiry, 0, nil, 200)
		if !strings.Contains(response.Body.String(), `id="modal"`) || strings.Contains(response.Body.String(), "<!DOCTYPE html>") {
			t.Fatalf("htmx=%t rendered non-modal edit response: %s", htmx, response.Body.String())
		}
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

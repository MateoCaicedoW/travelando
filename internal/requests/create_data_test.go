package requests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestInquiryFormRequiresValidGuestCedula(t *testing.T) {
	values := url.Values{
		"stay_type": {"day_use"}, "date_from": {"2099-02-10"}, "date_to": {"2099-02-10"}, "guests": {"2"},
		"guest_name": {"Huésped"}, "guest_phone": {"573001234567"}, "guest_email": {"guest@example.com"}, "terms_accepted": {"on"},
	}

	if errors := inquiryFormErrors(requestWithValues(values)); !errors.Has("guest_cedula") {
		t.Fatal("missing guest cédula should be rejected")
	}

	values.Set("guest_cedula", "12-34")
	if errors := inquiryFormErrors(requestWithValues(values)); !errors.Has("guest_cedula") {
		t.Fatal("short guest cédula should be rejected")
	}

	values.Set("guest_cedula", "1.234.567")
	if errors := inquiryFormErrors(requestWithValues(values)); errors.Has("guest_cedula") {
		t.Fatalf("formatted valid cédula should be accepted: %v", errors.ErrorStringFor("guest_cedula"))
	}
}

func TestInquiryFormRequiresTermsAcceptance(t *testing.T) {
	values := url.Values{
		"stay_type": {"day_use"}, "date_from": {"2099-02-10"}, "date_to": {"2099-02-10"}, "guests": {"2"},
		"guest_name": {"Huésped"}, "guest_cedula": {"12345678"}, "guest_phone": {"573001234567"}, "guest_email": {"guest@example.com"},
	}

	if errors := inquiryFormErrors(requestWithValues(values)); !errors.Has("terms_accepted") {
		t.Fatal("request without terms acceptance should be rejected")
	}
	values.Set("terms_accepted", "yes")
	if errors := inquiryFormErrors(requestWithValues(values)); !errors.Has("terms_accepted") {
		t.Fatal("unexpected checkbox value should be rejected")
	}
	values.Set("terms_accepted", "on")
	if errors := inquiryFormErrors(requestWithValues(values)); errors.Has("terms_accepted") {
		t.Fatalf("accepted terms should be valid: %v", errors.ErrorStringFor("terms_accepted"))
	}
}

func requestWithValues(values url.Values) *http.Request {
	r := httptest.NewRequest("POST", "/", nil)
	r.Form = values
	r.PostForm = values
	return r
}

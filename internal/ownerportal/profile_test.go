package ownerportal

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestOwnerProfileRequiresCurrentPINAndMatchingConfirmation(t *testing.T) {
	values := url.Values{
		"name": {"María López"}, "cedula": {"12345678"}, "whatsapp": {"573001234567"},
		"new_pin": {"654321"},
	}
	request := ownerProfileTestRequest(t, values)
	errors := ownerProfileErrors(request)
	if errors.ErrorStringFor("current_pin") == "" || errors.ErrorStringFor("confirm_pin") == "" {
		t.Fatalf("PIN change errors = %+v, want current PIN and confirmation errors", errors)
	}

	values.Set("current_pin", "123456")
	values.Set("confirm_pin", "111111")
	request = ownerProfileTestRequest(t, values)
	if message := ownerProfileErrors(request).ErrorStringFor("confirm_pin"); message == "" {
		t.Fatal("mismatched PIN confirmation was accepted")
	}

	values.Set("confirm_pin", "654321")
	request = ownerProfileTestRequest(t, values)
	if errors := ownerProfileErrors(request); errors.HasAny() {
		t.Fatalf("valid profile and PIN change rejected: %+v", errors)
	}
}

func TestOwnerProfileAllowsPersonalDataOnlyAndRejectsBlankName(t *testing.T) {
	values := url.Values{"name": {"María López"}, "cedula": {"12345678"}, "whatsapp": {"573001234567"}}
	if errors := ownerProfileErrors(ownerProfileTestRequest(t, values)); errors.HasAny() {
		t.Fatalf("profile data without PIN change rejected: %+v", errors)
	}

	values.Set("name", "   ")
	if message := ownerProfileErrors(ownerProfileTestRequest(t, values)).ErrorStringFor("name"); message == "" {
		t.Fatal("whitespace-only name was accepted")
	}
}

func TestOwnerProfileViewDoesNotRepublishPINValues(t *testing.T) {
	owner := models.Owner{ID: 7, Name: "María López", Cedula: "12345678", WhatsApp: "573001234567"}
	values := url.Values{"name": {owner.Name}, "cedula": {owner.Cedula}, "whatsapp": {owner.WhatsApp}, "current_pin": {"123456"}}
	var output bytes.Buffer
	if err := ownerProfileEl(owner, &ownerProfileState{Values: values}).Render(&output); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	for _, want := range []string{"Perfil del propietario", owner.Name, owner.Cedula, owner.WhatsApp, "PIN actual", "Nuevo PIN", "Confirmar PIN", `action="/owner/profile"`} {
		if !strings.Contains(html, want) {
			t.Errorf("owner profile view missing %q: %s", want, html)
		}
	}
	if strings.Contains(html, `value="123456"`) {
		t.Fatal("owner profile view republished submitted PIN")
	}
}

func ownerProfileTestRequest(t *testing.T, values url.Values) *http.Request {
	t.Helper()
	request := httptest.NewRequest("POST", "/owner/profile", strings.NewReader(values.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := request.ParseForm(); err != nil {
		t.Fatal(err)
	}
	return request
}

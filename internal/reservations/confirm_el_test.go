package reservations

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestConfirmationFormUsesScopedRouteAndPrefillsEstimate(t *testing.T) {
	inquiry := models.Inquiry{
		ID: 42, CabinName: "Cabaña", GuestName: "Huésped", StayType: models.StayDayUse,
		DateFrom: "2099-08-10", Estimated: 125000, DepositCOP: 20000,
	}
	for _, test := range []struct {
		owner bool
		path  string
	}{
		{path: "/admin/reservations/42/confirm"},
		{owner: true, path: "/owner/reservations/42/confirm"},
	} {
		var html strings.Builder
		if err := confirmReservationDialog(inquiry, test.owner, "125000", "").Render(&html); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{`action="` + test.path + `"`, `hx-post="` + test.path + `"`, `name="final_total"`, `value="125000"`, "Huésped", "Depósito incluido", "Estimado actual", "Total acordado (COP)", "Confirmar reserva"} {
			if !strings.Contains(html.String(), want) {
				t.Errorf("confirmation dialog missing %q: %s", want, html.String())
			}
		}
	}
}

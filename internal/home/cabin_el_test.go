package home

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestCabinDetailShowsGlobalDepositOrNoDeposit(t *testing.T) {
	for _, test := range []struct {
		amount int64
		want   string
	}{
		{amount: 250000, want: "$250.000 COP"},
		{amount: 0, want: "Sin depósito"},
	} {
		var html strings.Builder
		data := PageData{StayType: models.StayDayUse, Guests: 2, DepositCOP: test.amount}
		if err := CabinEl(models.Cabin{Name: "Cabaña", IncludedGuests: 2, AcceptsDayUse: true, DayPrice: 100000}, data).Render(&html); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(html.String(), "Depósito") || !strings.Contains(html.String(), test.want) {
			t.Errorf("cabin detail missing deposit %q: %s", test.want, html.String())
		}
		if test.amount > 0 && (!strings.Contains(html.String(), `data-price-estimate-deposit-cop-value="250000"`) || !strings.Contains(html.String(), "$350.000")) {
			t.Errorf("cabin estimate should include deposit: %s", html.String())
		}
	}
}

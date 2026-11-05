package dashboard

import (
	"bytes"
	"strings"
	"testing"

	"travelando/internal/system/services"
)

func TestParseDepositAmount(t *testing.T) {
	for _, test := range []struct {
		value string
		want  int64
		valid bool
	}{
		{value: "0", want: 0, valid: true},
		{value: "250000", want: 250000, valid: true},
		{value: "1000000000000", want: services.MaxAmountCOP, valid: true},
		{value: "-1"},
		{value: "1000000000001"},
		{value: "1.5"},
	} {
		got, err := parseDepositAmount(test.value)
		if test.valid && (err != nil || got != test.want) {
			t.Errorf("parseDepositAmount(%q) = %d, %v; want %d", test.value, got, err, test.want)
		}
		if !test.valid && err == nil {
			t.Errorf("parseDepositAmount(%q) should fail", test.value)
		}
	}
}

func TestDepositSettingIsAvailableInAdminConfiguration(t *testing.T) {
	var output bytes.Buffer
	if err := dashboardEl(nil, 0, 250000, nil, nil, 1).Render(&output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Depósito fijo", "se suma al estimado", `action="/admin/deposit"`, `value="250000"`, "Guardar depósito"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("admin deposit setting missing %q: %s", want, output.String())
		}
	}
}

package components

import (
	"bytes"
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestCabinCardShowsScheduleForSelectedStayType(t *testing.T) {
	cabin := models.Cabin{
		Name: "Cabaña", DayStart: "08:00", DayEnd: "18:00", CheckIn: "15:00", CheckOut: "12:00",
	}
	for _, test := range []struct {
		stayType models.StayType
		want     string
	}{
		{models.StayDayUse, "08:00 a 18:00"},
		{models.StayOvernight, "Entrada 15:00 · salida 12:00"},
	} {
		var output bytes.Buffer
		if err := CabinCard(cabin, 100000, test.stayType, "/cabins/1").Render(&output); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), test.want) {
			t.Errorf("%s cabin card missing schedule %q: %s", test.stayType, test.want, output.String())
		}
	}
}

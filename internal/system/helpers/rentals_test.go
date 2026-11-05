package helpers

import (
	"testing"

	"travelando/internal/system/models"
)

func TestScheduleLabelUsesStayType(t *testing.T) {
	if got, want := ScheduleLabel(models.StayDayUse, "08:00", "18:00"), "08:00 a 18:00"; got != want {
		t.Errorf("day-use schedule = %q, want %q", got, want)
	}
	if got, want := ScheduleLabel(models.StayOvernight, "15:00", "12:00"), "Entrada 15:00 · salida 12:00"; got != want {
		t.Errorf("overnight schedule = %q, want %q", got, want)
	}
}

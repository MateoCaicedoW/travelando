package services

import (
	"testing"

	"travelando/internal/system/models"
)

func TestReservationWindowUsesCabinScheduleAndCleaningTime(t *testing.T) {
	cabin := models.Cabin{DayStart: "08:00", DayEnd: "18:00", CheckIn: "15:00", CheckOut: "12:00", CleaningHours: 3}

	start, readyAt, err := reservationWindow(cabin, models.StayDayUse, "2026-11-10", "2026-11-10")
	if err != nil {
		t.Fatal(err)
	}
	if got := start.Format("2006-01-02 15:04"); got != "2026-11-10 08:00" {
		t.Errorf("day-use starts at %s, want 2026-11-10 08:00", got)
	}
	if got := readyAt.Format("2006-01-02 15:04"); got != "2026-11-10 21:00" {
		t.Errorf("day-use is ready at %s, want 2026-11-10 21:00", got)
	}

	start, readyAt, err = reservationWindow(cabin, models.StayOvernight, "2026-11-10", "2026-11-12")
	if err != nil {
		t.Fatal(err)
	}
	if got := start.Format("2006-01-02 15:04"); got != "2026-11-10 15:00" {
		t.Errorf("overnight starts at %s, want 2026-11-10 15:00", got)
	}
	if got := readyAt.Format("2006-01-02 15:04"); got != "2026-11-12 15:00" {
		t.Errorf("overnight is ready at %s, want 2026-11-12 15:00", got)
	}
}

func TestReservationWindowEndsAtCleaningCompletion(t *testing.T) {
	cabin := models.Cabin{DayStart: "08:00", DayEnd: "18:00", CheckIn: "21:00", CheckOut: "12:00", CleaningHours: 3}
	_, dayReady, err := reservationWindow(cabin, models.StayDayUse, "2026-11-10", "2026-11-10")
	if err != nil {
		t.Fatal(err)
	}
	nightStart, _, err := reservationWindow(cabin, models.StayOvernight, "2026-11-10", "2026-11-11")
	if err != nil {
		t.Fatal(err)
	}
	if !dayReady.Equal(nightStart) {
		t.Fatalf("next reservation starts at %v, want cleaning completion %v", nightStart, dayReady)
	}

	cabin.CheckIn = "20:59"
	nightStart, _, err = reservationWindow(cabin, models.StayOvernight, "2026-11-10", "2026-11-11")
	if err != nil {
		t.Fatal(err)
	}
	if !nightStart.Before(dayReady) {
		t.Fatalf("reservation starting at %v should precede cleaning completion %v", nightStart, dayReady)
	}
}

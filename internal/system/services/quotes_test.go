package services

import (
	"testing"
	"time"

	"travelando/internal/system/models"
)

func TestQuoteChargesOnlyGuestsBeyondIncludedCount(t *testing.T) {
	date := time.Date(2027, time.February, 10, 0, 0, 0, 0, time.UTC)
	base := QuoteInput{
		Type: models.StayDayUse, StartDate: date, EndDate: date,
		IncludedGuests: 20, ExtraGuestPrice: 15000, DayPrice: 300000,
	}
	for _, test := range []struct {
		guests int
		want   int64
	}{
		{guests: 20, want: 300000},
		{guests: 25, want: 375000},
	} {
		base.Guests = test.guests
		got, err := Quote(base)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("Quote(%d guests) = %d, want %d", test.guests, got, test.want)
		}
	}
}

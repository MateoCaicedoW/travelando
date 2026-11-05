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

func TestAddDepositToQuoteRejectsTotalsAboveCOPLimit(t *testing.T) {
	for _, test := range []struct {
		quote, deposit int64
		want           int64
		valid          bool
	}{
		{quote: 100000, deposit: 50000, want: 150000, valid: true},
		{quote: 0, deposit: 0, want: 0, valid: true},
		{quote: MaxAmountCOP, deposit: 1},
		{quote: 100, deposit: -1},
	} {
		got, err := AddDepositToQuote(test.quote, test.deposit)
		if test.valid && (err != nil || got != test.want) {
			t.Errorf("AddDepositToQuote(%d, %d) = %d, %v; want %d", test.quote, test.deposit, got, err, test.want)
		}
		if !test.valid && err == nil {
			t.Errorf("AddDepositToQuote(%d, %d) should reject invalid sum", test.quote, test.deposit)
		}
	}
}

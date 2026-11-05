package services

import (
	"time"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
)

const MaxAmountCOP int64 = 1_000_000_000_000

type QuoteInput struct {
	Type            models.StayType
	StartDate       time.Time
	EndDate         time.Time
	Guests          int
	IncludedGuests  int
	ExtraGuestPrice int64
	DayPrice        int64
	NightPrice      int64
}

const (
	StayDayUse    = models.StayDayUse
	StayOvernight = models.StayOvernight
)

func Quote(in QuoteInput) (int64, error) {
	if in.Guests < 1 || in.IncludedGuests < 1 || in.DayPrice < 0 || in.NightPrice < 0 || in.ExtraGuestPrice < 0 {
		return 0, ErrInvalidInput
	}

	unitPrice := in.DayPrice
	if in.Type == StayOvernight {
		unitPrice = in.NightPrice
	}
	if unitPrice > MaxAmountCOP || in.ExtraGuestPrice > MaxAmountCOP {
		return 0, ErrInvalidInput
	}

	units := int64(1)
	switch in.Type {
	case StayDayUse:
		if !helpers.SameDate(in.StartDate, in.EndDate) {
			return 0, ErrInvalidInput
		}
	case StayOvernight:
		if !in.EndDate.After(in.StartDate) {
			return 0, ErrInvalidInput
		}
		units = int64(in.EndDate.Sub(in.StartDate).Hours() / 24)
		if units < 1 || units > 365 {
			return 0, ErrInvalidInput
		}
	default:
		return 0, ErrInvalidInput
	}
	if unitPrice > MaxAmountCOP/units {
		return 0, ErrInvalidInput
	}
	total := unitPrice * units
	if in.Guests > in.IncludedGuests && in.ExtraGuestPrice > 0 {
		extraGuests := int64(in.Guests - in.IncludedGuests)
		if extraGuests > (MaxAmountCOP-total)/in.ExtraGuestPrice {
			return 0, ErrInvalidInput
		}
		total += extraGuests * in.ExtraGuestPrice
	}
	return total, nil
}

func Commission(total int64, basisPoints int) (int64, error) {
	if total < 0 || basisPoints < 0 || basisPoints > 10000 {
		return 0, ErrInvalidInput
	}
	whole := total / 10000 * int64(basisPoints)
	remainder := (total%10000*int64(basisPoints) + 5000) / 10000
	return whole + remainder, nil
}

package home

import (
	"time"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

func priceFor(cabin models.Cabin, stayType models.StayType, dateFrom, dateTo string, guests int, depositCOP int64) (int64, bool) {
	from, err := time.Parse(time.DateOnly, dateFrom)
	if err != nil {
		from = time.Now()
	}
	to, err := time.Parse(time.DateOnly, dateTo)
	if err != nil || stayType == models.StayDayUse && dateTo == "" {
		to = from
		if stayType == models.StayOvernight {
			to = from.AddDate(0, 0, 1)
		}
	}
	if guests < 1 {
		guests = 1
	}
	total, err := services.Quote(services.QuoteInput{
		Type: stayType, StartDate: from, EndDate: to, Guests: guests,
		IncludedGuests: cabin.IncludedGuests, ExtraGuestPrice: cabin.ExtraGuestPrice,
		DayPrice: cabin.DayPrice, NightPrice: cabin.NightPrice,
	})
	if err != nil {
		if stayType == models.StayDayUse {
			total = cabin.DayPrice
		} else {
			total = cabin.NightPrice
		}
	}
	total, err = services.AddDepositToQuote(total, depositCOP)
	return total, err == nil
}

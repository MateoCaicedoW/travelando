package home

import (
	"net/url"
	"strconv"

	"travelando/internal/system/models"
)

func CabinQueryData(query url.Values) PageData {
	data := PageData{StayType: models.StayDayUse, Guests: 2}
	if query.Get("stay_type") == string(models.StayOvernight) {
		data.StayType = models.StayOvernight
	}
	data.DateFrom = query.Get("date_from")
	data.DateTo = query.Get("date_to")
	if guests, err := strconv.Atoi(query.Get("guests")); err == nil && guests > 0 {
		data.Guests = guests
	}
	return data
}

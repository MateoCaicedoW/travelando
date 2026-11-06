package calendar

import (
	"errors"
	"net/http"
	"strconv"
	"time"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/system/services"
)

func ownerCalendarViewFromRequest(r *http.Request) (ownerCalendarView, error) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Calendar == nil {
		return ownerCalendarView{}, errors.New("calendar service unavailable")
	}
	now := time.Now()
	year, err := ownerCalendarParam(r, "year", now.Year())
	if err != nil {
		return ownerCalendarView{}, services.ErrInvalidInput
	}
	monthNumber, err := ownerCalendarParam(r, "month", int(now.Month()))
	if err != nil || monthNumber < 1 || monthNumber > 12 {
		return ownerCalendarView{}, services.ErrInvalidInput
	}
	data, err := set.Calendar.OwnerCalendar(auth.OwnerIDFromSession(r), year, time.Month(monthNumber))
	if err != nil {
		return ownerCalendarView{}, err
	}
	return newOwnerCalendarView(year, time.Month(monthNumber), data, now), nil
}

func ownerCalendarParam(r *http.Request, key string, fallback int) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

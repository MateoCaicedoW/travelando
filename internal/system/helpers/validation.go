package helpers

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

func ParseForm(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	return r.ParseForm()
}

func DigitsOnly(value string) string {
	var result strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func ValidTime(value string) bool {
	if len(value) != 5 || value[2] != ':' {
		return false
	}
	_, err := time.Parse("15:04", value)
	return err == nil
}

func ValidExternalURL(value string) bool {
	if value == "" {
		return true
	}
	u, err := url.ParseRequestURI(value)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && u.User == nil
}

func ValidTimeRange(start, end string) bool {
	startTime, startErr := time.Parse("15:04", start)
	endTime, endErr := time.Parse("15:04", end)
	return startErr == nil && endErr == nil && endTime.After(startTime)
}

func SameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

package helpers

import (
	"time"

	"travelando/internal/system/models"
)

func StayTypeLabel(stayType models.StayType) string {
	if stayType == models.StayOvernight {
		return "Estadía"
	}
	return "Pasadía"
}

func DateLabel(stayType models.StayType, dateFrom, dateTo string) string {
	from := formatDate(dateFrom)
	if stayType != models.StayOvernight {
		return from
	}
	return from + " – " + formatDate(dateTo)
}

func formatDate(value string) string {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return date.Format("02/01/2006")
}

package services

import (
	"time"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
)

type AvailabilityService interface {
	Search(SearchInput) (CabinSearchResult, error)
}

type SearchInput struct {
	StayType models.StayType
	DateFrom time.Time
	DateTo   time.Time
	Limit    int
	Offset   int
}

type CabinSearchResult struct {
	Cabins []models.Cabin
	Total  int
}

func (s *availabilityService) Search(input SearchInput) (CabinSearchResult, error) {
	if input.Limit < 0 || input.Offset < 0 || (input.StayType != models.StayDayUse && input.StayType != models.StayOvernight) {
		return CabinSearchResult{}, ErrInvalidInput
	}
	end := input.DateTo
	if input.StayType == models.StayDayUse {
		if !helpers.SameDate(input.DateFrom, input.DateTo) {
			return CabinSearchResult{}, ErrInvalidInput
		}
		end = input.DateFrom.AddDate(0, 0, 1)
	} else if !input.DateTo.After(input.DateFrom) {
		return CabinSearchResult{}, ErrInvalidInput
	}

	db, err := s.conn()
	if err != nil {
		return CabinSearchResult{}, err
	}
	query := cabinSearchSelect + `
		WHERE c.active = 1 AND o.active = 1 AND o.whatsapp <> ''
		  AND NOT EXISTS (
			SELECT 1 FROM inquiries i
			WHERE i.cabin_id = c.id AND i.status = 'reserved'
			  AND i.date_from < ? AND i.date_to > ?
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM cabin_blocks b
			WHERE b.cabin_id = c.id AND b.date_from < ? AND b.date_to > ?
		  )
		ORDER BY c.name, c.id
	`
	args := []any{end.Format("2006-01-02"), input.DateFrom.Format("2006-01-02"), end.Format("2006-01-02"), input.DateFrom.Format("2006-01-02")}
	if input.Limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, input.Limit, input.Offset)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return CabinSearchResult{}, err
	}
	defer rows.Close()

	result := CabinSearchResult{}
	for rows.Next() {
		var cabin models.Cabin
		var total int
		args := append(cabinScanArgs(&cabin), &total)
		if err := rows.Scan(args...); err != nil {
			return CabinSearchResult{}, err
		}
		result.Cabins = append(result.Cabins, cabin)
		result.Total = total
	}
	return result, rows.Err()
}

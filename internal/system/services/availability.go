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
	bookingEndDate := input.DateTo
	if input.StayType == models.StayDayUse {
		if !helpers.SameDate(input.DateFrom, input.DateTo) {
			return CabinSearchResult{}, ErrInvalidInput
		}
		end = input.DateFrom.AddDate(0, 0, 1)
		bookingEndDate = input.DateFrom
	} else if !input.DateTo.After(input.DateFrom) {
		return CabinSearchResult{}, ErrInvalidInput
	}

	db, err := s.conn()
	if err != nil {
		return CabinSearchResult{}, err
	}
	query := cabinSearchSelect + `
		WHERE c.active = 1 AND o.active = 1 AND o.whatsapp <> ''
		  AND `
	if input.StayType == models.StayDayUse {
		query += `c.accepts_day_use = 1`
	} else {
		query += `c.accepts_overnight = 1`
	}
	query += `
		  AND NOT EXISTS (
			SELECT 1 FROM inquiries i
			WHERE i.cabin_id = c.id AND i.status = 'reserved'
			  AND datetime(? || ' ' || CASE WHEN ? = 'day_use' THEN c.day_start ELSE c.check_in END) < datetime(
				  (CASE WHEN i.stay_type = 'day_use' THEN i.date_from ELSE i.date_to END) || ' ' || i.end_time,
				  printf('+%d hours', i.cleaning_hours)
			  )
			  AND datetime(
				  ? || ' ' || CASE WHEN ? = 'day_use' THEN c.day_end ELSE c.check_out END,
				  printf('+%d hours', c.cleaning_hours)
			  ) > datetime(i.date_from || ' ' || i.start_time)
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM cabin_blocks b
			WHERE b.cabin_id = c.id AND b.date_from < ? AND b.date_to > ?
		  )
		ORDER BY c.name, c.id
	`
	args := []any{
		input.DateFrom.Format(time.DateOnly), input.StayType, bookingEndDate.Format(time.DateOnly), input.StayType,
		end.Format(time.DateOnly), input.DateFrom.Format(time.DateOnly),
	}
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
	if err := rows.Err(); err != nil {
		return CabinSearchResult{}, err
	}
	if err := rows.Close(); err != nil {
		return CabinSearchResult{}, err
	}
	cabinIDs := make([]int64, len(result.Cabins))
	for i, cabin := range result.Cabins {
		cabinIDs[i] = cabin.ID
	}
	amenitiesByCabin, err := cabinAmenitiesFor(db, cabinIDs)
	if err != nil {
		return CabinSearchResult{}, err
	}
	for i := range result.Cabins {
		result.Cabins[i].Amenities = amenitiesByCabin[result.Cabins[i].ID]
	}
	return result, nil
}

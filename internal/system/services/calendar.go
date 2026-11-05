package services

import (
	"time"

	"travelando/internal/system/models"
)

type CalendarService interface {
	Calendar(year int, month time.Month) (CalendarData, error)
	OwnerCalendar(ownerID int64, year int, month time.Month) (CalendarData, error)
}

type CalendarData struct {
	OccupiedCabins []models.Cabin
	Entries        []CalendarEntry
}

type CalendarEntry struct {
	CabinID     int64
	CabinName   string
	InquiryID   int64
	DateFrom    string
	DateTo      string
	GuestName   string
	GuestPhone  string
	GuestEmail  string
	Guests      int
	StayType    models.StayType
	BlockReason string
	BlockSource string
	IsBlock     bool
	BlockID     int64
	StartAt     string
	EndAt       string
	ReadyAt     string
}

func (entry CalendarEntry) OccupiedRange() (time.Time, time.Time, bool) {
	if entry.IsBlock {
		from, errFrom := time.Parse(time.DateOnly, entry.DateFrom)
		to, errTo := time.Parse(time.DateOnly, entry.DateTo)
		return from, to, errFrom == nil && errTo == nil
	}
	start, errStart := time.Parse(time.DateTime, entry.StartAt)
	if errStart != nil {
		start, errStart = time.Parse("2006-01-02 15:04", entry.StartAt)
	}
	readyAt, errReady := time.Parse(time.DateTime, entry.ReadyAt)
	if errStart != nil || errReady != nil {
		return time.Time{}, time.Time{}, false
	}
	return start, readyAt, true
}

func (s *calendarService) Calendar(year int, month time.Month) (CalendarData, error) {
	if year < 1 || year > 9998 || month < time.January || month > time.December {
		return CalendarData{}, ErrInvalidInput
	}

	db, err := s.conn()
	if err != nil {
		return CalendarData{}, err
	}
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	end := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	reservationReady := `datetime((CASE WHEN i.stay_type = 'day_use' THEN i.date_from ELSE i.date_to END) || ' ' || i.end_time, printf('+%d hours', i.cleaning_hours))`
	data := CalendarData{}
	cabins, err := db.Query(cabinSelect+`
		WHERE EXISTS (
			SELECT 1 FROM inquiries i
			WHERE i.cabin_id = c.id AND i.status = 'reserved' AND i.date_from < ? AND `+reservationReady+` > datetime(?)
		) OR EXISTS (
			SELECT 1 FROM cabin_blocks b
			WHERE b.cabin_id = c.id AND b.date_from < ? AND b.date_to > ?
		)
		ORDER BY c.active DESC, c.name, c.id
	`, end, start, end, start)
	if err != nil {
		return CalendarData{}, err
	}
	for cabins.Next() {
		cabin, err := scanCabin(cabins)
		if err != nil {
			cabins.Close()
			return CalendarData{}, err
		}
		data.OccupiedCabins = append(data.OccupiedCabins, cabin)
	}
	if err := cabins.Err(); err != nil {
		cabins.Close()
		return CalendarData{}, err
	}
	if err := cabins.Close(); err != nil {
		return CalendarData{}, err
	}
	rows, err := db.Query(`
		SELECT i.cabin_id, c.name, i.id, i.date_from, i.date_to, i.guest_name, i.guest_phone, i.guest_email, i.guests, i.stay_type, '', '', 0, 0,
		       i.date_from || ' ' || i.start_time,
		       (CASE WHEN i.stay_type = 'day_use' THEN i.date_from ELSE i.date_to END) || ' ' || i.end_time,
		       datetime((CASE WHEN i.stay_type = 'day_use' THEN i.date_from ELSE i.date_to END) || ' ' || i.end_time, printf('+%d hours', i.cleaning_hours))
		FROM inquiries i JOIN cabins c ON c.id = i.cabin_id
		WHERE i.status = 'reserved' AND i.date_from < ? AND `+reservationReady+` > datetime(?)
		UNION ALL
		SELECT b.cabin_id, c.name, 0, b.date_from, b.date_to, '', '', '', 0, '', b.reason, b.source, 1, b.id, '', '', ''
		FROM cabin_blocks b JOIN cabins c ON c.id = b.cabin_id
		WHERE b.date_from < ? AND b.date_to > ?
		ORDER BY 3, 2
	`, end, start, end, start)
	if err != nil {
		return CalendarData{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry CalendarEntry
		if err := rows.Scan(&entry.CabinID, &entry.CabinName, &entry.InquiryID, &entry.DateFrom, &entry.DateTo, &entry.GuestName, &entry.GuestPhone, &entry.GuestEmail, &entry.Guests, &entry.StayType, &entry.BlockReason, &entry.BlockSource, &entry.IsBlock, &entry.BlockID, &entry.StartAt, &entry.EndAt, &entry.ReadyAt); err != nil {
			return CalendarData{}, err
		}
		data.Entries = append(data.Entries, entry)
	}
	return data, rows.Err()
}

func (s *calendarService) OwnerCalendar(ownerID int64, year int, month time.Month) (CalendarData, error) {
	if ownerID < 1 || year < 1 || year > 9998 || month < time.January || month > time.December {
		return CalendarData{}, ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return CalendarData{}, err
	}
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	end := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	reservationReady := `datetime((CASE WHEN i.stay_type = 'day_use' THEN i.date_from ELSE i.date_to END) || ' ' || i.end_time, printf('+%d hours', i.cleaning_hours))`
	data := CalendarData{}
	cabins, err := db.Query(cabinSelect+` WHERE c.owner_id = ? AND o.active = 1 ORDER BY c.name, c.id`, ownerID)
	if err != nil {
		return CalendarData{}, err
	}
	for cabins.Next() {
		cabin, err := scanCabin(cabins)
		if err != nil {
			cabins.Close()
			return CalendarData{}, err
		}
		data.OccupiedCabins = append(data.OccupiedCabins, cabin)
	}
	if err := cabins.Err(); err != nil {
		cabins.Close()
		return CalendarData{}, err
	}
	if err := cabins.Close(); err != nil {
		return CalendarData{}, err
	}
	rows, err := db.Query(`
		SELECT i.cabin_id, c.name, i.id, i.date_from, i.date_to, i.guest_name, i.guest_phone, i.guest_email, i.guests, i.stay_type, '', '', 0, 0,
		       i.date_from || ' ' || i.start_time,
		       (CASE WHEN i.stay_type = 'day_use' THEN i.date_from ELSE i.date_to END) || ' ' || i.end_time,
	       `+reservationReady+`
		FROM inquiries i JOIN cabins c ON c.id = i.cabin_id JOIN owners o ON o.id = c.owner_id
		WHERE c.owner_id = ? AND o.active = 1 AND i.status = 'reserved' AND i.date_from < ? AND `+reservationReady+` > datetime(?)
		UNION ALL
		SELECT b.cabin_id, c.name, 0, b.date_from, b.date_to, '', '', '', 0, '', '', b.source, 1, b.id, '', '', ''
		FROM cabin_blocks b JOIN cabins c ON c.id = b.cabin_id JOIN owners o ON o.id = c.owner_id
		WHERE c.owner_id = ? AND o.active = 1 AND b.date_from < ? AND b.date_to > ?
		ORDER BY 3, 2
	`, ownerID, end, start, ownerID, end, start)
	if err != nil {
		return CalendarData{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry CalendarEntry
		if err := rows.Scan(&entry.CabinID, &entry.CabinName, &entry.InquiryID, &entry.DateFrom, &entry.DateTo, &entry.GuestName, &entry.GuestPhone, &entry.GuestEmail, &entry.Guests, &entry.StayType, &entry.BlockReason, &entry.BlockSource, &entry.IsBlock, &entry.BlockID, &entry.StartAt, &entry.EndAt, &entry.ReadyAt); err != nil {
			return CalendarData{}, err
		}
		data.Entries = append(data.Entries, entry)
	}
	return data, rows.Err()
}

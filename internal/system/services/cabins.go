package services

import (
	"database/sql"
	"errors"
	"strings"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
)

type CabinService interface {
	CabinsPage(page, pageSize int) (CabinPageResult, error)
	SearchCabinsPage(search string, page, pageSize int) (CabinPageResult, error)
	Cabin(id int64) (*models.Cabin, error)
	CreateCabin(*models.Cabin) error
	UpdateCabin(*models.Cabin) error
	CreateCabinForOwner(ownerID int64, cabin *models.Cabin) error
	UpdateCabinForOwner(ownerID int64, cabin *models.Cabin) error
}

type CabinPageResult struct {
	Cabins     []models.Cabin
	Page       int
	Total      int
	TotalPages int
}

func (s *cabinService) CabinsPage(page, pageSize int) (CabinPageResult, error) {
	return s.SearchCabinsPage("", page, pageSize)
}

func (s *cabinService) SearchCabinsPage(search string, page, pageSize int) (CabinPageResult, error) {
	if page < 1 || pageSize < 1 {
		return CabinPageResult{}, ErrInvalidInput
	}
	where, args := cabinSearchQuery(search)
	db, err := s.conn()
	if err != nil {
		return CabinPageResult{}, err
	}
	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM cabins c LEFT JOIN owners o ON o.id = c.owner_id`+where, args...).Scan(&total); err != nil {
		return CabinPageResult{}, err
	}
	totalPages := max(1, (total+pageSize-1)/pageSize)
	page = min(page, totalPages)
	queryArgs := append(args, pageSize, (page-1)*pageSize)
	rows, err := db.Query(cabinSelect+where+` ORDER BY c.active DESC, c.name, c.id LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return CabinPageResult{}, err
	}
	defer rows.Close()

	result := CabinPageResult{Page: page, Total: total, TotalPages: totalPages}
	for rows.Next() {
		cabin, err := scanCabin(rows)
		if err != nil {
			return CabinPageResult{}, err
		}
		result.Cabins = append(result.Cabins, cabin)
	}
	return result, rows.Err()
}

func cabinSearchQuery(search string) (string, []any) {
	search = strings.TrimSpace(search)
	if search == "" {
		return "", nil
	}
	pattern := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(search) + "%"
	return ` WHERE (c.name LIKE ? ESCAPE '!' OR c.location LIKE ? ESCAPE '!' OR o.name LIKE ? ESCAPE '!' OR o.whatsapp LIKE ? ESCAPE '!')`, []any{pattern, pattern, pattern, pattern}
}

func (s *cabinService) Cabin(id int64) (*models.Cabin, error) {
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	cabin, err := scanCabin(db.QueryRow(cabinSelect+` WHERE c.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &cabin, nil
}

func validateCabin(cabin *models.Cabin) error {
	cabin.Name = strings.TrimSpace(cabin.Name)
	cabin.Description = strings.TrimSpace(cabin.Description)
	cabin.Location = strings.TrimSpace(cabin.Location)
	cabin.LocationURL = strings.TrimSpace(cabin.LocationURL)
	cabin.CoverImage = strings.TrimSpace(cabin.CoverImage)
	if cabin.DayStart == "" {
		cabin.DayStart = "09:00"
	}
	if cabin.DayEnd == "" {
		cabin.DayEnd = "18:00"
	}
	if cabin.CheckIn == "" {
		cabin.CheckIn = "15:00"
	}
	if cabin.CheckOut == "" {
		cabin.CheckOut = "12:00"
	}
	if cabin.Name == "" || len(cabin.Name) > 120 || len(cabin.Description) > 2000 || len(cabin.Location) > 200 || len(cabin.LocationURL) > 2048 || len(cabin.CoverImage) > 2048 || cabin.IncludedGuests < 1 || cabin.DayPrice < 0 || cabin.DayPrice > MaxAmountCOP || cabin.NightPrice < 0 || cabin.NightPrice > MaxAmountCOP || cabin.ExtraGuestPrice < 0 || cabin.ExtraGuestPrice > MaxAmountCOP || cabin.CleaningHours < 0 || cabin.CleaningHours > 24 || !helpers.ValidTime(cabin.DayStart) || !helpers.ValidTime(cabin.DayEnd) || !helpers.ValidTime(cabin.CheckIn) || !helpers.ValidTime(cabin.CheckOut) || !helpers.ValidTimeRange(cabin.DayStart, cabin.DayEnd) || !helpers.ValidExternalURL(cabin.LocationURL) || !helpers.ValidExternalURL(cabin.CoverImage) {
		return ErrInvalidInput
	}
	return nil
}

func (s *cabinService) CreateCabin(cabin *models.Cabin) error {
	if cabin.ID != 0 {
		return ErrInvalidInput
	}
	if err := validateCabin(cabin); err != nil {
		return err
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	if err := validateCabinOwner(db, cabin.OwnerID); err != nil {
		return err
	}
	result, err := db.Exec(`
		INSERT INTO cabins (name, description, location, location_url, cover_image, included_guests, extra_guest_price, day_price, night_price, accepts_day_use, accepts_overnight, day_start, day_end, check_in, check_out, cleaning_hours, active, owner_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cabin.Name, cabin.Description, cabin.Location, cabin.LocationURL, cabin.CoverImage, cabin.IncludedGuests, cabin.ExtraGuestPrice, cabin.DayPrice, cabin.NightPrice, cabin.AcceptsDayUse, cabin.AcceptsOvernight, cabin.DayStart, cabin.DayEnd, cabin.CheckIn, cabin.CheckOut, cabin.CleaningHours, cabin.Active, cabin.OwnerID)
	if err != nil {
		return err
	}
	cabin.ID, err = result.LastInsertId()
	return err
}

func (s *cabinService) CreateCabinForOwner(ownerID int64, cabin *models.Cabin) error {
	if ownerID < 1 || cabin == nil || cabin.OwnerID != 0 && cabin.OwnerID != ownerID {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	if err := validateActiveCabinOwner(db, ownerID); err != nil {
		return err
	}
	cabin.OwnerID = ownerID
	return s.CreateCabin(cabin)
}

func (s *cabinService) UpdateCabin(cabin *models.Cabin) error {
	if cabin.ID < 1 {
		return ErrInvalidInput
	}
	if err := validateCabin(cabin); err != nil {
		return err
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	if err := validateCabinOwner(db, cabin.OwnerID); err != nil {
		return err
	}
	result, err := db.Exec(`
		UPDATE cabins
		SET name = ?, description = ?, location = ?, location_url = ?, cover_image = ?, included_guests = ?, extra_guest_price = ?, day_price = ?, night_price = ?, accepts_day_use = ?, accepts_overnight = ?, day_start = ?, day_end = ?, check_in = ?, check_out = ?, cleaning_hours = ?, active = ?, owner_id = ?
		WHERE id = ?
	`, cabin.Name, cabin.Description, cabin.Location, cabin.LocationURL, cabin.CoverImage, cabin.IncludedGuests, cabin.ExtraGuestPrice, cabin.DayPrice, cabin.NightPrice, cabin.AcceptsDayUse, cabin.AcceptsOvernight, cabin.DayStart, cabin.DayEnd, cabin.CheckIn, cabin.CheckOut, cabin.CleaningHours, cabin.Active, cabin.OwnerID, cabin.ID)
	if err != nil {
		return err
	}
	return requireUpdated(result)
}

func (s *cabinService) UpdateCabinForOwner(ownerID int64, cabin *models.Cabin) error {
	if ownerID < 1 || cabin == nil || cabin.ID < 1 || cabin.OwnerID != 0 && cabin.OwnerID != ownerID {
		return ErrInvalidInput
	}
	cabin.OwnerID = ownerID
	if err := validateCabin(cabin); err != nil {
		return err
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	if err := validateCabinOwner(db, ownerID); err != nil {
		return err
	}
	result, err := db.Exec(`
		UPDATE cabins
		SET name = ?, description = ?, location = ?, location_url = ?, cover_image = ?, included_guests = ?, extra_guest_price = ?, day_price = ?, night_price = ?, accepts_day_use = ?, accepts_overnight = ?, day_start = ?, day_end = ?, check_in = ?, check_out = ?, cleaning_hours = ?, active = ?
		WHERE id = ? AND owner_id = ? AND EXISTS (SELECT 1 FROM owners WHERE id = ? AND active = 1)
	`, cabin.Name, cabin.Description, cabin.Location, cabin.LocationURL, cabin.CoverImage, cabin.IncludedGuests, cabin.ExtraGuestPrice, cabin.DayPrice, cabin.NightPrice, cabin.AcceptsDayUse, cabin.AcceptsOvernight, cabin.DayStart, cabin.DayEnd, cabin.CheckIn, cabin.CheckOut, cabin.CleaningHours, cabin.Active, cabin.ID, ownerID, ownerID)
	if err != nil {
		return err
	}
	return requireUpdated(result)
}

func validateCabinOwner(db *sql.DB, ownerID int64) error {
	if ownerID < 1 {
		return ErrInvalidInput
	}
	var exists int
	if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM owners WHERE id = ?)`, ownerID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return ErrInvalidInput
	}
	return nil
}

func validateActiveCabinOwner(db *sql.DB, ownerID int64) error {
	var exists int
	if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM owners WHERE id = ? AND active = 1)`, ownerID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return ErrNotFound
	}
	return nil
}

const cabinSelect = `
	SELECT ` + cabinSelectColumns + `
	FROM cabins c LEFT JOIN owners o ON o.id = c.owner_id
`

const cabinSearchSelect = `
	SELECT ` + cabinSelectColumns + `, COUNT(*) OVER()
	FROM cabins c LEFT JOIN owners o ON o.id = c.owner_id
`

const cabinSelectColumns = `
	c.id, COALESCE(c.owner_id, 0), COALESCE(o.name, ''), COALESCE(o.whatsapp, ''), COALESCE(o.active, 0), c.name, c.description, c.location, c.location_url, c.cover_image,
    c.included_guests, c.extra_guest_price, c.day_price, c.night_price, c.accepts_day_use, c.accepts_overnight, c.day_start, c.day_end, c.check_in, c.check_out, c.cleaning_hours, c.active
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCabin(row rowScanner) (models.Cabin, error) {
	var cabin models.Cabin
	err := row.Scan(cabinScanArgs(&cabin)...)
	return cabin, err
}

func cabinScanArgs(cabin *models.Cabin) []any {
	return []any{&cabin.ID, &cabin.OwnerID, &cabin.OwnerName, &cabin.OwnerWhatsApp, &cabin.OwnerActive, &cabin.Name, &cabin.Description, &cabin.Location, &cabin.LocationURL, &cabin.CoverImage, &cabin.IncludedGuests, &cabin.ExtraGuestPrice, &cabin.DayPrice, &cabin.NightPrice, &cabin.AcceptsDayUse, &cabin.AcceptsOvernight, &cabin.DayStart, &cabin.DayEnd, &cabin.CheckIn, &cabin.CheckOut, &cabin.CleaningHours, &cabin.Active}
}

func requireUpdated(result sql.Result) error {
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

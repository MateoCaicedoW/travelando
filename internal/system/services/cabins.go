package services

import (
	"database/sql"
	"errors"
	"strings"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
)

type CabinService interface {
	FeaturedCabin() (*models.Cabin, error)
	RandomPublicCabins() ([]models.Cabin, error)
	CabinsPage(page, pageSize int) (CabinPageResult, error)
	SearchCabinsPage(search string, page, pageSize int) (CabinPageResult, error)
	Cabin(id int64) (*models.Cabin, error)
	CreateCabin(*models.Cabin) error
	UpdateCabin(*models.Cabin) error
	AddCabinPhoto(cabinID int64, photoURL string) error
	DeleteCabinPhoto(cabinID int64, photoURL string) error
	CreateCabinForOwner(ownerID int64, cabin *models.Cabin) error
	UpdateCabinForOwner(ownerID int64, cabin *models.Cabin) error
}

type CabinPageResult struct {
	Cabins     []models.Cabin
	Page       int
	Total      int
	TotalPages int
}

func (s *cabinService) FeaturedCabin() (*models.Cabin, error) {
	db, err := s.conn()
	if err != nil {
		return nil, err
	}

	var cabin models.Cabin
	var photo models.CabinPhoto
	err = db.QueryRow(`
		SELECT ` + cabinSelectColumns + `, p.photo_url, p.position
		FROM cabins c
		JOIN owners o ON o.id = c.owner_id
		JOIN cabin_photos p ON p.cabin_id = c.id
		WHERE c.active = 1 AND o.active = 1 AND trim(o.whatsapp) <> '' AND trim(p.photo_url) <> ''
		ORDER BY c.id, p.position, p.id
		LIMIT 1
	`).Scan(append(cabinScanArgs(&cabin), &photo.URL, &photo.Position)...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	cabin.Photos = []models.CabinPhoto{photo}
	return &cabin, nil
}

func (s *cabinService) RandomPublicCabins() ([]models.Cabin, error) {
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(cabinSelect + `
		WHERE c.active = 1 AND o.active = 1 AND trim(o.whatsapp) <> ''
		  AND (c.accepts_day_use = 1 OR c.accepts_overnight = 1)
		ORDER BY RANDOM()
		LIMIT 3
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cabins []models.Cabin
	for rows.Next() {
		cabin, err := scanCabin(rows)
		if err != nil {
			return nil, err
		}
		cabins = append(cabins, cabin)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return cabins, attachCabinPhotos(db, cabins)
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
	if err := rows.Err(); err != nil {
		return CabinPageResult{}, err
	}
	if err := rows.Close(); err != nil {
		return CabinPageResult{}, err
	}
	return result, attachCabinPhotos(db, result.Cabins)
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
	cabin.Amenities, err = cabinAmenities(db, cabin.ID)
	if err != nil {
		return nil, err
	}
	photos, err := cabinPhotosFor(db, []int64{cabin.ID})
	if err != nil {
		return nil, err
	}
	cabin.Photos = photos[cabin.ID]
	return &cabin, nil
}

func validateCabin(cabin *models.Cabin) error {
	cabin.Name = strings.TrimSpace(cabin.Name)
	cabin.Description = strings.TrimSpace(cabin.Description)
	cabin.Location = strings.TrimSpace(cabin.Location)
	cabin.LocationURL = strings.TrimSpace(cabin.LocationURL)
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
	if cabin.Name == "" || len(cabin.Name) > 120 || len(cabin.Description) > 2000 || len(cabin.Location) > 200 || len(cabin.LocationURL) > 2048 || cabin.IncludedGuests < 1 || cabin.DayPrice < 0 || cabin.DayPrice > MaxAmountCOP || cabin.NightPrice < 0 || cabin.NightPrice > MaxAmountCOP || cabin.ExtraGuestPrice < 0 || cabin.ExtraGuestPrice > MaxAmountCOP || cabin.CleaningHours < 0 || cabin.CleaningHours > 24 || !helpers.ValidTime(cabin.DayStart) || !helpers.ValidTime(cabin.DayEnd) || !helpers.ValidTime(cabin.CheckIn) || !helpers.ValidTime(cabin.CheckOut) || !helpers.ValidTimeRange(cabin.DayStart, cabin.DayEnd) || !helpers.ValidExternalURL(cabin.LocationURL) {
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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`
		INSERT INTO cabins (name, description, location, location_url, included_guests, extra_guest_price, day_price, night_price, accepts_day_use, accepts_overnight, day_start, day_end, check_in, check_out, cleaning_hours, active, owner_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cabin.Name, cabin.Description, cabin.Location, cabin.LocationURL, cabin.IncludedGuests, cabin.ExtraGuestPrice, cabin.DayPrice, cabin.NightPrice, cabin.AcceptsDayUse, cabin.AcceptsOvernight, cabin.DayStart, cabin.DayEnd, cabin.CheckIn, cabin.CheckOut, cabin.CleaningHours, cabin.Active, cabin.OwnerID)
	if err != nil {
		return err
	}
	cabin.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}
	if err := saveCabinAmenities(tx, cabin.ID, cabin.OwnerID, cabin.Amenities); err != nil {
		return err
	}
	return tx.Commit()
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

func (s *cabinService) AddCabinPhoto(cabinID int64, photoURL string) error {
	db, err := s.conn()
	if err != nil {
		return err
	}
	var pos int
	_ = db.QueryRow(`SELECT COALESCE(MAX(position), -1) + 1 FROM cabin_photos WHERE cabin_id = ?`, cabinID).Scan(&pos)
	_, err = db.Exec(`INSERT INTO cabin_photos (cabin_id, photo_url, position) VALUES (?, ?, ?)`, cabinID, photoURL, pos)
	return err
}

func (s *cabinService) DeleteCabinPhoto(cabinID int64, photoURL string) error {
	db, err := s.conn()
	if err != nil {
		return err
	}
	_, err = db.Exec(`DELETE FROM cabin_photos WHERE cabin_id = ? AND photo_url = ?`, cabinID, photoURL)
	return err
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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`
		UPDATE cabins
		SET name = ?, description = ?, location = ?, location_url = ?, included_guests = ?, extra_guest_price = ?, day_price = ?, night_price = ?, accepts_day_use = ?, accepts_overnight = ?, day_start = ?, day_end = ?, check_in = ?, check_out = ?, cleaning_hours = ?, active = ?, owner_id = ?
		WHERE id = ?
	`, cabin.Name, cabin.Description, cabin.Location, cabin.LocationURL, cabin.IncludedGuests, cabin.ExtraGuestPrice, cabin.DayPrice, cabin.NightPrice, cabin.AcceptsDayUse, cabin.AcceptsOvernight, cabin.DayStart, cabin.DayEnd, cabin.CheckIn, cabin.CheckOut, cabin.CleaningHours, cabin.Active, cabin.OwnerID, cabin.ID)
	if err != nil {
		return err
	}
	if err := requireUpdated(result); err != nil {
		return err
	}
	if err := saveCabinAmenities(tx, cabin.ID, cabin.OwnerID, cabin.Amenities); err != nil {
		return err
	}
	return tx.Commit()
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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`
		UPDATE cabins
		SET name = ?, description = ?, location = ?, location_url = ?, included_guests = ?, extra_guest_price = ?, day_price = ?, night_price = ?, accepts_day_use = ?, accepts_overnight = ?, day_start = ?, day_end = ?, check_in = ?, check_out = ?, cleaning_hours = ?, active = ?
		WHERE id = ? AND owner_id = ? AND EXISTS (SELECT 1 FROM owners WHERE id = ? AND active = 1)
	`, cabin.Name, cabin.Description, cabin.Location, cabin.LocationURL, cabin.IncludedGuests, cabin.ExtraGuestPrice, cabin.DayPrice, cabin.NightPrice, cabin.AcceptsDayUse, cabin.AcceptsOvernight, cabin.DayStart, cabin.DayEnd, cabin.CheckIn, cabin.CheckOut, cabin.CleaningHours, cabin.Active, cabin.ID, ownerID, ownerID)
	if err != nil {
		return err
	}
	if err := requireUpdated(result); err != nil {
		return err
	}
	if err := saveCabinAmenities(tx, cabin.ID, ownerID, cabin.Amenities); err != nil {
		return err
	}
	return tx.Commit()
}

type cabinAmenityTransaction interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
}

func saveCabinAmenities(tx cabinAmenityTransaction, cabinID, ownerID int64, amenities []models.CabinAmenity) error {
	if _, err := tx.Exec(`DELETE FROM cabin_amenities WHERE cabin_id = ?`, cabinID); err != nil {
		return err
	}
	seen := make(map[int64]struct{}, len(amenities))
	for _, amenity := range amenities {
		if amenity.TypeID < 1 {
			return ErrInvalidInput
		}
		if _, ok := seen[amenity.TypeID]; ok {
			return ErrInvalidInput
		}
		seen[amenity.TypeID] = struct{}{}
		var typeOwner sql.NullInt64
		var valueType models.AmenityValueType
		if err := tx.QueryRow(`SELECT owner_id, value_type FROM cabin_amenity_types WHERE id = ?`, amenity.TypeID).Scan(&typeOwner, &valueType); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrInvalidInput
			}
			return err
		}
		if typeOwner.Valid && typeOwner.Int64 != ownerID {
			return ErrInvalidInput
		}
		var boolValue, quantityValue any
		switch valueType {
		case models.AmenityBoolean:
			boolValue = amenity.BooleanValue
		case models.AmenityQuantity:
			if amenity.QuantityValue < 0 {
				return ErrInvalidInput
			}
			quantityValue = amenity.QuantityValue
		default:
			return ErrInvalidInput
		}
		if _, err := tx.Exec(`INSERT INTO cabin_amenities (cabin_id, amenity_type_id, boolean_value, quantity_value) VALUES (?, ?, ?, ?)`, cabinID, amenity.TypeID, boolValue, quantityValue); err != nil {
			return err
		}
	}
	return nil
}

func cabinAmenities(db *sql.DB, cabinID int64) ([]models.CabinAmenity, error) {
	amenitiesByCabin, err := cabinAmenitiesFor(db, []int64{cabinID})
	return amenitiesByCabin[cabinID], err
}

func cabinAmenitiesFor(db *sql.DB, cabinIDs []int64) (map[int64][]models.CabinAmenity, error) {
	amenitiesByCabin := make(map[int64][]models.CabinAmenity, len(cabinIDs))
	if len(cabinIDs) == 0 {
		return amenitiesByCabin, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(cabinIDs)), ",")
	args := make([]any, len(cabinIDs))
	for i, id := range cabinIDs {
		args[i] = id
	}
	rows, err := db.Query(`
		SELECT a.cabin_id, t.id, t.name, t.value_type, a.boolean_value, a.quantity_value
		FROM cabin_amenities a
		JOIN cabin_amenity_types t ON t.id = a.amenity_type_id
		WHERE a.cabin_id IN (`+placeholders+`)
		ORDER BY a.cabin_id, t.name COLLATE NOCASE, t.id
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cabinID int64
		var amenity models.CabinAmenity
		var booleanValue, quantityValue sql.NullInt64
		if err := rows.Scan(&cabinID, &amenity.TypeID, &amenity.Name, &amenity.ValueType, &booleanValue, &quantityValue); err != nil {
			return nil, err
		}
		amenity.BooleanValue = booleanValue.Valid && booleanValue.Int64 == 1
		if quantityValue.Valid {
			amenity.QuantityValue = int(quantityValue.Int64)
		}
		amenitiesByCabin[cabinID] = append(amenitiesByCabin[cabinID], amenity)
	}
	return amenitiesByCabin, rows.Err()
}

func cabinPhotosFor(db *sql.DB, cabinIDs []int64) (map[int64][]models.CabinPhoto, error) {
	photosByCabin := make(map[int64][]models.CabinPhoto, len(cabinIDs))
	if len(cabinIDs) == 0 {
		return photosByCabin, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(cabinIDs)), ",")
	args := make([]any, len(cabinIDs))
	for i, id := range cabinIDs {
		args[i] = id
	}
	rows, err := db.Query(`
		SELECT cabin_id, photo_url, position
		FROM cabin_photos
		WHERE cabin_id IN (`+placeholders+`)
		ORDER BY cabin_id, position, id
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cabinID int64
		var photo models.CabinPhoto
		if err := rows.Scan(&cabinID, &photo.URL, &photo.Position); err != nil {
			return nil, err
		}
		photosByCabin[cabinID] = append(photosByCabin[cabinID], photo)
	}
	return photosByCabin, rows.Err()
}

func attachCabinPhotos(db *sql.DB, cabins []models.Cabin) error {
	cabinIDs := make([]int64, len(cabins))
	for i := range cabins {
		cabinIDs[i] = cabins[i].ID
	}
	photosByCabin, err := cabinPhotosFor(db, cabinIDs)
	if err != nil {
		return err
	}
	for i := range cabins {
		cabins[i].Photos = photosByCabin[cabins[i].ID]
	}
	return nil
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
	c.id, COALESCE(c.owner_id, 0), COALESCE(o.name, ''), COALESCE(o.whatsapp, ''), COALESCE(o.active, 0), c.name, c.description, c.location, c.location_url,
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
	return []any{&cabin.ID, &cabin.OwnerID, &cabin.OwnerName, &cabin.OwnerWhatsApp, &cabin.OwnerActive, &cabin.Name, &cabin.Description, &cabin.Location, &cabin.LocationURL, &cabin.IncludedGuests, &cabin.ExtraGuestPrice, &cabin.DayPrice, &cabin.NightPrice, &cabin.AcceptsDayUse, &cabin.AcceptsOvernight, &cabin.DayStart, &cabin.DayEnd, &cabin.CheckIn, &cabin.CheckOut, &cabin.CleaningHours, &cabin.Active}
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

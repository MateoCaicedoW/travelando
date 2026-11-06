package services

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"unicode/utf8"
)

const (
	ownerPINIterations = 600_000
	ownerPINLength     = 6
	ownerLockoutAfter  = 5
	ownerLockoutFor    = 15 * time.Minute
)

type OwnerService interface {
	Owners() ([]models.Owner, error)
	OwnerByID(ownerID int64) (*models.Owner, error)
	CreateOwner(*models.Owner, string) error
	UpdateOwner(*models.Owner, string) error
	AuthenticateOwner(cedula, pin string) (*models.Owner, error)
	OwnerCabins(ownerID int64) ([]models.Cabin, error)
	OwnerCabinsPage(ownerID int64, page, pageSize int) (CabinPageResult, error)
	OwnerSessionValid(ownerID, sessionVersion int64) (bool, error)
	SetOwnerActive(ownerID int64, active bool) error
}

func (s *ownerService) Owners() ([]models.Owner, error) {
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT o.id, o.name, o.cedula, COALESCE(o.whatsapp, ''), COUNT(DISTINCT c.id),
			COALESCE(SUM(i.commission_paid_amount), 0),
			COALESCE(SUM(MAX(COALESCE(i.commission_due, 0) - i.commission_paid_amount, 0)), 0),
			COALESCE(SUM(MAX(i.commission_paid_amount - COALESCE(i.commission_due, 0), 0)), 0),
			CASE WHEN o.pin_hash <> '' THEN 1 ELSE 0 END, o.active
		FROM owners o
		LEFT JOIN cabins c ON c.owner_id = o.id
		LEFT JOIN inquiries i ON i.cabin_id = c.id AND (i.status = 'reserved' OR i.cancelled_at IS NOT NULL)
		GROUP BY o.id
		ORDER BY o.name, o.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var owners []models.Owner
	for rows.Next() {
		var owner models.Owner
		var hasPIN, active int
		if err := rows.Scan(&owner.ID, &owner.Name, &owner.Cedula, &owner.WhatsApp, &owner.CabinCount, &owner.CommissionPaid, &owner.CommissionDue, &owner.CommissionCredit, &hasPIN, &active); err != nil {
			return nil, err
		}
		owner.HasPIN = hasPIN != 0
		owner.Active = active != 0
		owners = append(owners, owner)
	}
	return owners, rows.Err()
}

func (s *ownerService) OwnerByID(ownerID int64) (*models.Owner, error) {
	if ownerID < 1 {
		return nil, ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	var owner models.Owner
	err = db.QueryRow(`SELECT id, name, cedula, COALESCE(whatsapp, '') FROM owners WHERE id = ?`, ownerID).
		Scan(&owner.ID, &owner.Name, &owner.Cedula, &owner.WhatsApp)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &owner, nil
}

func (s *ownerService) CreateOwner(owner *models.Owner, pin string) error {
	if owner == nil || owner.ID != 0 || !validOwnerPIN(pin) || validateOwner(owner, pin) != nil {
		return ErrInvalidInput
	}
	hash, err := hashOwnerPIN(pin)
	if err != nil {
		return err
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
whatsapp := any(owner.WhatsApp)
	if owner.WhatsApp == "" {
		whatsapp = nil
	}
	result, err := db.Exec(`INSERT INTO owners (name, cedula, whatsapp, pin_hash) VALUES (?, ?, ?, ?)`, owner.Name, owner.Cedula, whatsapp, hash)
	if err != nil {
		if ownerConflict(err) {
			return ErrConflict
		}
		return err
	}
	owner.ID, err = result.LastInsertId()
	return err
}

func (s *ownerService) UpdateOwner(owner *models.Owner, pin string) error {
	if owner == nil || owner.ID < 1 || validateOwner(owner, pin) != nil {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	var currentHash string
	if pin == "" {
		if err := db.QueryRow(`SELECT pin_hash FROM owners WHERE id = ?`, owner.ID).Scan(&currentHash); errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		} else if err != nil {
			return err
		}
		if currentHash == "" {
			return ErrInvalidInput
		}
	}
	var result sql.Result
	whatsapp := any(owner.WhatsApp)
	if owner.WhatsApp == "" {
		whatsapp = nil
	}
	if pin == "" {
		result, err = db.Exec(`UPDATE owners SET name = ?, cedula = ?, whatsapp = ?, auth_version = auth_version + CASE WHEN cedula <> ? OR whatsapp <> ? THEN 1 ELSE 0 END WHERE id = ?`, owner.Name, owner.Cedula, whatsapp, owner.Cedula, whatsapp, owner.ID)
	} else {
		var hash string
		hash, err = hashOwnerPIN(pin)
		if err == nil {
			result, err = db.Exec(`UPDATE owners SET name = ?, cedula = ?, whatsapp = ?, pin_hash = ?, failed_attempts = 0, locked_until = 0, auth_version = auth_version + 1 WHERE id = ?`, owner.Name, owner.Cedula, whatsapp, hash, owner.ID)
		}
	}
	if err != nil {
		if ownerConflict(err) {
			return ErrConflict
		}
		return err
	}
	return requireUpdated(result)
}

func (s *ownerService) AuthenticateOwner(cedula, pin string) (*models.Owner, error) {
	cedula = helpers.DigitsOnly(cedula)
	if !validCedula(cedula) || !validOwnerPIN(pin) {
		return nil, ErrUnauthorized
	}
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	var owner models.Owner
	var pinHash string
	var lockedUntil int64
	err = db.QueryRow(`SELECT id, name, cedula, COALESCE(whatsapp, ''), pin_hash, locked_until, auth_version FROM owners WHERE cedula = ? AND cedula <> '' AND active = 1`, cedula).
		Scan(&owner.ID, &owner.Name, &owner.Cedula, &owner.WhatsApp, &pinHash, &lockedUntil, &owner.SessionVersion)
	if errors.Is(err, sql.ErrNoRows) {
		verifyOwnerPIN(pin, unknownOwnerPINHash())
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	if lockedUntil > now || !verifyOwnerPIN(pin, pinHash) {
		_, err := db.Exec(`UPDATE owners SET
			failed_attempts = failed_attempts + 1,
			locked_until = CASE WHEN failed_attempts + 1 >= ? THEN ? ELSE locked_until END
			WHERE id = ? AND locked_until <= ?`, ownerLockoutAfter, now+int64(ownerLockoutFor.Seconds()), owner.ID, now)
		if err != nil {
			return nil, err
		}
		return nil, ErrUnauthorized
	}
	if _, err := db.Exec(`UPDATE owners SET failed_attempts = 0, locked_until = 0 WHERE id = ?`, owner.ID); err != nil {
		return nil, err
	}
	return &owner, nil
}

func (s *ownerService) OwnerCabins(ownerID int64) ([]models.Cabin, error) {
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(cabinSelect+` WHERE c.owner_id = ? ORDER BY c.name, c.id`, ownerID)
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
	return cabins, rows.Err()
}

func (s *ownerService) OwnerCabinsPage(ownerID int64, page, pageSize int) (CabinPageResult, error) {
	if ownerID < 1 || page < 1 || pageSize < 1 {
		return CabinPageResult{}, ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return CabinPageResult{}, err
	}
	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM cabins WHERE owner_id = ?`, ownerID).Scan(&total); err != nil {
		return CabinPageResult{}, err
	}
	totalPages := max(1, (total+pageSize-1)/pageSize)
	page = min(page, totalPages)
	rows, err := db.Query(cabinSelect+` WHERE c.owner_id = ? ORDER BY c.active DESC, c.name, c.id LIMIT ? OFFSET ?`, ownerID, pageSize, (page-1)*pageSize)
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

func (s *ownerService) OwnerSessionValid(ownerID, sessionVersion int64) (bool, error) {
	if ownerID < 1 || sessionVersion < 1 {
		return false, nil
	}
	db, err := s.conn()
	if err != nil {
		return false, err
	}
	var valid int
	err = db.QueryRow(`SELECT EXISTS (SELECT 1 FROM owners WHERE id = ? AND auth_version = ? AND active = 1)`, ownerID, sessionVersion).Scan(&valid)
	return valid == 1, err
}

func (s *ownerService) SetOwnerActive(ownerID int64, active bool) error {
	if ownerID < 1 {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	activeInt := 0
	if active {
		activeInt = 1
	}
	result, err := db.Exec(`UPDATE owners SET active = ?, auth_version = auth_version + 1 WHERE id = ? AND active <> ?`, activeInt, ownerID, activeInt)
	if err != nil {
		return err
	}
	if n, err := result.RowsAffected(); err != nil {
		return err
	} else if n > 0 {
		return nil
	}
	var exists int
	if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM owners WHERE id = ?)`, ownerID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return ErrNotFound
	}
	return nil
}

func validateOwner(owner *models.Owner, pin string) error {
	if owner == nil {
		return ErrInvalidInput
	}
	owner.Name = strings.TrimSpace(owner.Name)
	owner.Cedula = helpers.DigitsOnly(owner.Cedula)
	owner.WhatsApp = helpers.DigitsOnly(owner.WhatsApp)
	if owner.Name == "" || utf8.RuneCountInString(owner.Name) > 120 || !validCedula(owner.Cedula) || (owner.WhatsApp != "" && (len(owner.WhatsApp) < 10 || len(owner.WhatsApp) > 15)) || pin != "" && !validOwnerPIN(pin) {
		return ErrInvalidInput
	}
	return nil
}

func validCedula(cedula string) bool { return len(cedula) >= 5 && len(cedula) <= 15 }

func ownerConflict(err error) bool {
	return strings.Contains(err.Error(), "owners.whatsapp") || strings.Contains(err.Error(), "owners.cedula")
}

func validOwnerPIN(pin string) bool {
	if len(pin) != ownerPINLength {
		return false
	}
	for _, digit := range pin {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func hashOwnerPIN(pin string) (string, error) {
	var salt [16]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return "", err
	}
	derived, err := pbkdf2.Key(sha256.New, pin, salt[:], ownerPINIterations, sha256.Size)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", ownerPINIterations, base64.RawStdEncoding.EncodeToString(salt[:]), base64.RawStdEncoding.EncodeToString(derived)), nil
}

func verifyOwnerPIN(pin, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" || parts[1] != fmt.Sprint(ownerPINIterations) {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil || len(salt) < 16 {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(want) != sha256.Size {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, pin, salt, ownerPINIterations, len(want))
	return err == nil && subtle.ConstantTimeCompare(got, want) == 1
}

var unknownPINOnce sync.Once
var unknownPINHash string

func unknownOwnerPINHash() string {
	unknownPINOnce.Do(func() { unknownPINHash, _ = hashOwnerPIN("000000") })
	return unknownPINHash
}

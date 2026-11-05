package services

import (
	"database/sql"
	"errors"
	"strings"

	"travelando/internal/system/models"
)

type AmenityService interface {
	TypesForOwner(ownerID int64) ([]models.AmenityType, error)
	CreateType(ownerID int64, amenityType *models.AmenityType) error
	UpdateType(ownerID, typeID int64, amenityType models.AmenityType) error
	DeleteType(ownerID, typeID int64) error
}

type amenityService struct{ serviceDB }

func (s *amenityService) TypesForOwner(ownerID int64) ([]models.AmenityType, error) {
	if ownerID < 0 {
		return nil, ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT id, COALESCE(owner_id, 0), name, value_type
		FROM cabin_amenity_types
		WHERE owner_id IS NULL OR owner_id = ?
		ORDER BY owner_id IS NOT NULL, name COLLATE NOCASE, id
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []models.AmenityType
	for rows.Next() {
		var amenityType models.AmenityType
		if err := rows.Scan(&amenityType.ID, &amenityType.OwnerID, &amenityType.Name, &amenityType.ValueType); err != nil {
			return nil, err
		}
		types = append(types, amenityType)
	}
	return types, rows.Err()
}

func (s *amenityService) CreateType(ownerID int64, amenityType *models.AmenityType) error {
	if amenityType == nil || ownerID < 0 || !validAmenityType(amenityType) {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	if ownerID > 0 {
		if err := validateActiveCabinOwner(db, ownerID); err != nil {
			return err
		}
	}
	var owner any
	if ownerID > 0 {
		owner = ownerID
	}
	result, err := db.Exec(`INSERT INTO cabin_amenity_types (owner_id, name, value_type) VALUES (?, ?, ?)`, owner, amenityType.Name, amenityType.ValueType)
	if err != nil {
		if isUniqueConstraint(err) {
			return ErrInvalidInput
		}
		return err
	}
	amenityType.ID, err = result.LastInsertId()
	amenityType.OwnerID = ownerID
	return err
}

func (s *amenityService) UpdateType(ownerID, typeID int64, amenityType models.AmenityType) error {
	if ownerID < 0 || typeID < 1 || !validAmenityType(&amenityType) {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var oldType models.AmenityValueType
	if err := tx.QueryRow(`SELECT value_type FROM cabin_amenity_types WHERE id = ? AND owner_id IS ?`, typeID, nullableOwnerID(ownerID)).Scan(&oldType); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	result, err := tx.Exec(`UPDATE cabin_amenity_types SET name = ?, value_type = ? WHERE id = ? AND owner_id IS ?`, amenityType.Name, amenityType.ValueType, typeID, nullableOwnerID(ownerID))
	if err != nil {
		if isUniqueConstraint(err) {
			return ErrInvalidInput
		}
		return err
	}
	if oldType != amenityType.ValueType {
		if _, err := tx.Exec(`DELETE FROM cabin_amenities WHERE amenity_type_id = ?`, typeID); err != nil {
			return err
		}
	}
	if err := requireUpdated(result); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *amenityService) DeleteType(ownerID, typeID int64) error {
	if ownerID < 0 || typeID < 1 {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM cabin_amenity_types WHERE id = ? AND owner_id IS ?)`, typeID, nullableOwnerID(ownerID)).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	if _, err := tx.Exec(`DELETE FROM cabin_amenities WHERE amenity_type_id = ?`, typeID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM cabin_amenity_types WHERE id = ? AND owner_id IS ?`, typeID, nullableOwnerID(ownerID)); err != nil {
		return err
	}
	return tx.Commit()
}

func validAmenityType(amenityType *models.AmenityType) bool {
	amenityType.Name = strings.TrimSpace(amenityType.Name)
	return amenityType.Name != "" && len(amenityType.Name) <= 80 && (amenityType.ValueType == models.AmenityBoolean || amenityType.ValueType == models.AmenityQuantity)
}

func nullableOwnerID(ownerID int64) any {
	if ownerID == 0 {
		return nil
	}
	return ownerID
}

func isUniqueConstraint(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}

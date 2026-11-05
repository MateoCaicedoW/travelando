package services

import (
	"strings"
	"time"

	"travelando/internal/system/models"
)

type BlockService interface {
	Blocks(cabinID int64) ([]models.CabinBlock, error)
	AddBlock(*models.CabinBlock) error
	AddExternalBlock(ownerID int64, block *models.CabinBlock) error
	DeleteBlock(cabinID, blockID int64) error
	DeleteExternalBlock(ownerID, cabinID, blockID int64) error
}

func (s *blockService) Blocks(cabinID int64) ([]models.CabinBlock, error) {
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT id, cabin_id, date_from, date_to, reason, source
		FROM cabin_blocks WHERE cabin_id = ? ORDER BY date_from
	`, cabinID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []models.CabinBlock
	for rows.Next() {
		var block models.CabinBlock
		if err := rows.Scan(&block.ID, &block.CabinID, &block.DateFrom, &block.DateTo, &block.Reason, &block.Source); err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, rows.Err()
}

func (s *blockService) AddBlock(block *models.CabinBlock) error {
	if err := validateBlock(block); err != nil {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	result, err := db.Exec(`
		INSERT INTO cabin_blocks (cabin_id, date_from, date_to, reason)
		SELECT ?, ?, ?, ?
		WHERE NOT EXISTS (
			SELECT 1 FROM inquiries i
			WHERE i.cabin_id = ? AND i.status = 'reserved'
			  AND i.date_from < ? AND i.date_to > ?
		)
	`, block.CabinID, block.DateFrom, block.DateTo, block.Reason, block.CabinID, block.DateTo, block.DateFrom)
	if err != nil {
		if strings.Contains(err.Error(), "dates unavailable") {
			return ErrUnavailable
		}
		return err
	}
	if n, err := result.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrUnavailable
	}
	block.ID, err = result.LastInsertId()
	return err
}

func (s *blockService) AddExternalBlock(ownerID int64, block *models.CabinBlock) error {
	if ownerID < 1 || block == nil || block.CabinID < 1 || validateBlock(block) != nil {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	result, err := db.Exec(`
		INSERT INTO cabin_blocks (cabin_id, date_from, date_to, reason, source)
		SELECT ?, ?, ?, '', 'external'
		WHERE EXISTS (SELECT 1 FROM cabins c JOIN owners o ON o.id = c.owner_id WHERE c.id = ? AND c.owner_id = ? AND o.active = 1)
		  AND NOT EXISTS (
			SELECT 1 FROM inquiries i
			WHERE i.cabin_id = ? AND i.status = 'reserved'
			  AND i.date_from < ? AND i.date_to > ?
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM cabin_blocks b
			WHERE b.cabin_id = ? AND b.date_from < ? AND b.date_to > ?
		  )
	`, block.CabinID, block.DateFrom, block.DateTo, block.CabinID, ownerID,
		block.CabinID, block.DateTo, block.DateFrom, block.CabinID, block.DateTo, block.DateFrom)
	if err != nil {
		if strings.Contains(err.Error(), "dates unavailable") {
			return ErrUnavailable
		}
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var assigned int
		if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM cabins c JOIN owners o ON o.id = c.owner_id WHERE c.id = ? AND c.owner_id = ? AND o.active = 1)`, block.CabinID, ownerID).Scan(&assigned); err != nil {
			return err
		}
		if assigned == 0 {
			return ErrNotFound
		}
		return ErrUnavailable
	}
	block.Source = "external"
	block.Reason = ""
	block.ID, err = result.LastInsertId()
	return err
}

func validateBlock(block *models.CabinBlock) error {
	if block == nil {
		return ErrInvalidInput
	}
	from, err := time.Parse("2006-01-02", block.DateFrom)
	if err != nil {
		return ErrInvalidInput
	}
	to, err := time.Parse("2006-01-02", block.DateTo)
	if err != nil || !to.After(from) || len(block.Reason) > 200 {
		return ErrInvalidInput
	}
	block.Reason = strings.TrimSpace(block.Reason)
	return nil
}

func (s *blockService) DeleteBlock(cabinID, id int64) error {
	db, err := s.conn()
	if err != nil {
		return err
	}
	result, err := db.Exec(`DELETE FROM cabin_blocks WHERE id = ? AND cabin_id = ?`, id, cabinID)
	if err != nil {
		return err
	}
	if n, err := result.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *blockService) DeleteExternalBlock(ownerID, cabinID, id int64) error {
	if ownerID < 1 || cabinID < 1 || id < 1 {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	result, err := db.Exec(`
		DELETE FROM cabin_blocks
		WHERE id = ? AND cabin_id = ? AND source = 'external'
		  AND EXISTS (SELECT 1 FROM cabins c JOIN owners o ON o.id = c.owner_id WHERE c.id = cabin_blocks.cabin_id AND c.owner_id = ? AND o.active = 1)
	`, id, cabinID, ownerID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

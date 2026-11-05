package services

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
)

type InquiryService interface {
	CreateInquiry(*models.Inquiry) (string, error)
	Inquiries() ([]models.Inquiry, error)
	InquiryByToken(token string) (*models.Inquiry, error)
	ResolveInquiry(token string, status models.InquiryStatus, finalTotal int64) error
	MarkCommissionPaid(inquiryID int64) error
}

func (s *inquiryService) CreateInquiry(inquiry *models.Inquiry) (string, error) {
	inquiry.GuestName = strings.TrimSpace(inquiry.GuestName)
	inquiry.GuestPhone = helpers.DigitsOnly(inquiry.GuestPhone)
	inquiry.GuestEmail = strings.TrimSpace(inquiry.GuestEmail)
	from, err := time.Parse("2006-01-02", inquiry.DateFrom)
	if err != nil || inquiry.GuestName == "" || len(inquiry.GuestName) > 120 || len(inquiry.GuestPhone) < 8 || len(inquiry.GuestPhone) > 15 || len(inquiry.GuestEmail) > 254 || inquiry.Guests < 1 {
		return "", ErrInvalidInput
	}
	if inquiry.DateFrom < time.Now().Format("2006-01-02") {
		return "", ErrInvalidInput
	}
	address, err := mail.ParseAddress(inquiry.GuestEmail)
	if err != nil || address.Address != inquiry.GuestEmail {
		return "", ErrInvalidInput
	}
	to := from
	if inquiry.StayType == models.StayDayUse {
		if inquiry.DateTo != "" {
			to, err = time.Parse("2006-01-02", inquiry.DateTo)
			if err != nil {
				return "", ErrInvalidInput
			}
		}
		if !helpers.SameDate(from, to) {
			return "", ErrInvalidInput
		}
		inquiry.DateTo = inquiry.DateFrom
	} else {
		if inquiry.StayType != models.StayOvernight {
			return "", ErrInvalidInput
		}
		to, err = time.Parse("2006-01-02", inquiry.DateTo)
		if err != nil || !to.After(from) || to.Sub(from).Hours()/24 > 365 {
			return "", ErrInvalidInput
		}
	}

	db, err := s.conn()
	if err != nil {
		return "", err
	}
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	cabin, err := scanCabin(tx.QueryRow(cabinSelect+` WHERE c.id = ? AND c.active = 1 AND o.active = 1 AND c.owner_id > 0 AND o.whatsapp <> ''`, inquiry.CabinID))
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	quote, err := Quote(QuoteInput{
		Type: inquiry.StayType, StartDate: from, EndDate: to, Guests: inquiry.Guests,
		IncludedGuests: cabin.IncludedGuests, ExtraGuestPrice: cabin.ExtraGuestPrice,
		DayPrice: cabin.DayPrice, NightPrice: cabin.NightPrice,
	})
	if err != nil {
		return "", err
	}
	dateTo := inquiry.DateTo
	if inquiry.StayType == models.StayDayUse {
		dateTo = from.AddDate(0, 0, 1).Format("2006-01-02")
	}
	var occupied int
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM inquiries i
			WHERE i.cabin_id = ? AND i.status = 'reserved'
			  AND i.date_from < ? AND i.date_to > ?
		) OR EXISTS (
			SELECT 1 FROM cabin_blocks b
			WHERE b.cabin_id = ? AND b.date_from < ? AND b.date_to > ?
		)
	`, cabin.ID, dateTo, inquiry.DateFrom, cabin.ID, dateTo, inquiry.DateFrom).Scan(&occupied)
	if err != nil {
		return "", err
	}
	if occupied != 0 {
		return "", ErrUnavailable
	}

	token, err := newReportToken()
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(`
		INSERT INTO inquiries (cabin_id, guest_name, guest_phone, guest_email, stay_type, date_from, date_to, guests, estimated_total, report_token_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cabin.ID, inquiry.GuestName, inquiry.GuestPhone, inquiry.GuestEmail, inquiry.StayType, inquiry.DateFrom, dateTo, inquiry.Guests, quote, hashToken(token))
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}

	inquiry.CabinName = cabin.Name
	inquiry.OwnerName = cabin.OwnerName
	inquiry.WhatsApp = cabin.OwnerWhatsApp
	inquiry.Estimated = quote
	inquiry.Status = models.InquiryPending
	return token, nil
}

func (s *inquiryService) MarkCommissionPaid(inquiryID int64) error {
	if inquiryID < 1 {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	result, err := db.Exec(`UPDATE inquiries SET commission_paid_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'reserved' AND commission_due > 0 AND commission_paid_at IS NULL`, inquiryID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	var exists int
	if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM inquiries WHERE id = ?)`, inquiryID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return ErrNotFound
	}
	return ErrInvalidInput
}

func (s *inquiryService) Inquiries() ([]models.Inquiry, error) {
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(inquirySelect + ` ORDER BY i.created_at DESC, i.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inquiries []models.Inquiry
	for rows.Next() {
		inquiry, err := scanInquiry(rows)
		if err != nil {
			return nil, err
		}
		inquiries = append(inquiries, inquiry)
	}
	return inquiries, rows.Err()
}

func (s *inquiryService) InquiryByToken(token string) (*models.Inquiry, error) {
	if len(token) < 32 || len(token) > 64 {
		return nil, ErrNotFound
	}
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	inquiry, err := scanInquiry(db.QueryRow(inquirySelect+` WHERE i.report_token_hash = ?`, hashToken(token)))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &inquiry, nil
}

func (s *inquiryService) ResolveInquiry(token string, status models.InquiryStatus, finalTotal int64) error {
	if status != models.InquiryReserved && status != models.InquiryLost || status == models.InquiryReserved && (finalTotal < 0 || finalTotal > MaxAmountCOP) {
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

	var current models.InquiryStatus
	var commissionBP int
	err = tx.QueryRow(`
		SELECT i.status, cs.commission_bps
		FROM inquiries i JOIN commission_settings cs ON cs.id = 1
		WHERE i.report_token_hash = ?
	`, hashToken(token)).Scan(&current, &commissionBP)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if current != models.InquiryPending {
		return ErrInvalidInput
	}

	var result sql.Result
	if status == models.InquiryLost {
		result, err = tx.Exec(`UPDATE inquiries SET status = 'lost', resolved_at = CURRENT_TIMESTAMP WHERE report_token_hash = ? AND status = 'pending'`, hashToken(token))
	} else {
		commission, err := Commission(finalTotal, commissionBP)
		if err != nil {
			return err
		}
		result, err = tx.Exec(`
			UPDATE inquiries
			SET status = 'reserved', final_total = ?, commission_bps = ?, commission_due = ?, resolved_at = CURRENT_TIMESTAMP
			WHERE report_token_hash = ? AND status = 'pending'
			  AND NOT EXISTS (
				SELECT 1 FROM inquiries other
				WHERE other.cabin_id = inquiries.cabin_id AND other.id <> inquiries.id AND other.status = 'reserved'
				  AND other.date_from < inquiries.date_to AND other.date_to > inquiries.date_from
			  )
			  AND NOT EXISTS (
				SELECT 1 FROM cabin_blocks b
				WHERE b.cabin_id = inquiries.cabin_id
				  AND b.date_from < inquiries.date_to AND b.date_to > inquiries.date_from
			  )
		`, finalTotal, commissionBP, commission, hashToken(token))
	}
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
		if status == models.InquiryReserved {
			return ErrUnavailable
		}
		return ErrInvalidInput
	}
	return tx.Commit()
}

const inquirySelect = `
	SELECT i.id, i.cabin_id, c.name, COALESCE(o.name, ''), COALESCE(o.whatsapp, ''), i.guest_name, i.guest_phone, i.guest_email,
	       i.stay_type, i.date_from, i.date_to, i.guests, i.estimated_total, i.status,
	       COALESCE(i.final_total, 0), COALESCE(i.commission_bps, cs.commission_bps), COALESCE(i.commission_due, 0), COALESCE(i.commission_paid_at, ''), i.created_at
	FROM inquiries i JOIN cabins c ON c.id = i.cabin_id LEFT JOIN owners o ON o.id = c.owner_id JOIN commission_settings cs ON cs.id = 1
`

func scanInquiry(row rowScanner) (models.Inquiry, error) {
	var inquiry models.Inquiry
	err := row.Scan(
		&inquiry.ID, &inquiry.CabinID, &inquiry.CabinName, &inquiry.OwnerName, &inquiry.WhatsApp,
		&inquiry.GuestName, &inquiry.GuestPhone, &inquiry.GuestEmail, &inquiry.StayType, &inquiry.DateFrom, &inquiry.DateTo,
		&inquiry.Guests, &inquiry.Estimated, &inquiry.Status, &inquiry.FinalTotal, &inquiry.CommissionBP, &inquiry.Commission, &inquiry.CommissionPaidAt, &inquiry.CreatedAt,
	)
	return inquiry, err
}

func newReportToken() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

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
	ReservationsForOwner(ownerID int64) ([]models.Inquiry, error)
	InquiryByToken(token string) (*models.Inquiry, error)
	Reservation(id int64) (*models.Inquiry, error)
	AdminReservation(id int64) (*models.Inquiry, error)
	InquiryForOwner(ownerID, id int64) (*models.Inquiry, error)
	ReservationForOwner(ownerID, id int64) (*models.Inquiry, error)
	ResolveInquiry(token string, status models.InquiryStatus, finalTotal int64) error
	UpdatePendingInquiry(id int64, update models.InquiryUpdate) error
	UpdatePendingInquiryForOwner(ownerID, id int64, update models.InquiryUpdate) error
	UpdateReservation(id int64, update models.ReservationUpdate) error
	UpdateReservationForOwner(ownerID, id int64, update models.ReservationUpdate) error
	CancelReservation(id int64) error
	CancelReservationForOwner(ownerID, id int64) error
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
	var deposit int64
	if err := tx.QueryRow(`SELECT deposit_cop FROM commission_settings WHERE id = 1`).Scan(&deposit); err != nil {
		return "", err
	}
	if inquiry.StayType == models.StayDayUse && !cabin.AcceptsDayUse || inquiry.StayType == models.StayOvernight && !cabin.AcceptsOvernight {
		return "", ErrStayTypeNotAccepted
	}
	quote, err := Quote(QuoteInput{
		Type: inquiry.StayType, StartDate: from, EndDate: to, Guests: inquiry.Guests,
		IncludedGuests: cabin.IncludedGuests, ExtraGuestPrice: cabin.ExtraGuestPrice,
		DayPrice: cabin.DayPrice, NightPrice: cabin.NightPrice,
	})
	if err != nil {
		return "", err
	}
	estimated, err := AddDepositToQuote(quote, deposit)
	if err != nil {
		return "", err
	}
	dateTo := inquiry.DateTo
	if inquiry.StayType == models.StayDayUse {
		dateTo = from.AddDate(0, 0, 1).Format("2006-01-02")
		inquiry.StartTime, inquiry.EndTime = cabin.DayStart, cabin.DayEnd
	} else {
		inquiry.StartTime, inquiry.EndTime = cabin.CheckIn, cabin.CheckOut
	}
	inquiry.CleaningHours = cabin.CleaningHours
	startAt, readyAt, err := reservationWindow(cabin, inquiry.StayType, inquiry.DateFrom, inquiry.DateTo)
	if err != nil {
		return "", err
	}
	var occupied int
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM inquiries i
			WHERE i.cabin_id = ? AND i.status = 'reserved'
			  AND datetime(?) < datetime(
				  (CASE WHEN i.stay_type = 'day_use' THEN i.date_from ELSE i.date_to END) || ' ' || i.end_time,
				  printf('+%d hours', i.cleaning_hours)
			  )
			  AND datetime(?) > datetime(i.date_from || ' ' || i.start_time)
		) OR EXISTS (
			SELECT 1 FROM cabin_blocks b
			WHERE b.cabin_id = ? AND b.date_from < ? AND b.date_to > ?
		)
	`, cabin.ID, startAt.Format("2006-01-02 15:04:05"), readyAt.Format("2006-01-02 15:04:05"), cabin.ID, dateTo, inquiry.DateFrom).Scan(&occupied)
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
		INSERT INTO inquiries (cabin_id, guest_name, guest_phone, guest_email, stay_type, date_from, date_to, start_time, end_time, cleaning_hours, guests, estimated_total, deposit_cop, report_token_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cabin.ID, inquiry.GuestName, inquiry.GuestPhone, inquiry.GuestEmail, inquiry.StayType, inquiry.DateFrom, dateTo, inquiry.StartTime, inquiry.EndTime, inquiry.CleaningHours, inquiry.Guests, estimated, deposit, hashToken(token))
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}

	inquiry.CabinName = cabin.Name
	inquiry.OwnerName = cabin.OwnerName
	inquiry.WhatsApp = cabin.OwnerWhatsApp
	inquiry.Estimated = estimated
	inquiry.DepositCOP = deposit
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
	result, err := db.Exec(`UPDATE inquiries SET commission_paid_amount = commission_due, commission_paid_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'reserved' AND commission_due > commission_paid_amount`, inquiryID)
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

func (s *inquiryService) Reservation(inquiryID int64) (*models.Inquiry, error) {
	return s.reservation(inquiryID, 0)
}

func (s *inquiryService) AdminReservation(inquiryID int64) (*models.Inquiry, error) {
	if inquiryID < 1 {
		return nil, ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	inquiry, err := scanInquiry(db.QueryRow(inquirySelect+` WHERE i.id = ? AND (i.status IN ('pending', 'reserved') OR i.cancelled_at IS NOT NULL)`, inquiryID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	inquiry.CommissionAdjustments, err = commissionAdjustments(db, inquiryID)
	if err != nil {
		return nil, err
	}
	return &inquiry, nil
}

func (s *inquiryService) InquiryForOwner(ownerID, inquiryID int64) (*models.Inquiry, error) {
	if ownerID < 1 || inquiryID < 1 {
		return nil, ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	inquiry, err := scanInquiry(db.QueryRow(inquirySelect+` WHERE i.id = ? AND c.owner_id = ? AND (i.status IN ('pending', 'reserved') OR i.cancelled_at IS NOT NULL)`, inquiryID, ownerID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	inquiry.CommissionAdjustments, err = commissionAdjustments(db, inquiryID)
	if err != nil {
		return nil, err
	}
	return &inquiry, nil
}

func (s *inquiryService) ReservationForOwner(ownerID, inquiryID int64) (*models.Inquiry, error) {
	if ownerID < 1 {
		return nil, ErrInvalidInput
	}
	return s.reservation(inquiryID, ownerID)
}

func (s *inquiryService) reservation(inquiryID, ownerID int64) (*models.Inquiry, error) {
	if inquiryID < 1 {
		return nil, ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	query := inquirySelect + ` WHERE i.id = ? AND i.status = 'reserved'`
	args := []any{inquiryID}
	if ownerID > 0 {
		query += ` AND c.owner_id = ?`
		args = append(args, ownerID)
	}
	inquiry, err := scanInquiry(db.QueryRow(query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	inquiry.CommissionAdjustments, err = commissionAdjustments(db, inquiryID)
	if err != nil {
		return nil, err
	}
	return &inquiry, nil
}

func commissionAdjustments(db *sql.DB, inquiryID int64) ([]models.CommissionAdjustment, error) {
	rows, err := db.Query(`SELECT previous_amount, new_amount, amount, actor_role, created_at FROM commission_adjustments WHERE inquiry_id = ? ORDER BY id`, inquiryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var adjustments []models.CommissionAdjustment
	for rows.Next() {
		var adjustment models.CommissionAdjustment
		if err := rows.Scan(&adjustment.PreviousAmount, &adjustment.NewAmount, &adjustment.Amount, &adjustment.ActorRole, &adjustment.CreatedAt); err != nil {
			return nil, err
		}
		adjustments = append(adjustments, adjustment)
	}
	return adjustments, rows.Err()
}

func (s *inquiryService) UpdateReservation(inquiryID int64, update models.ReservationUpdate) error {
	return s.updateReservation(inquiryID, 0, "admin", update)
}

func (s *inquiryService) UpdateReservationForOwner(ownerID, inquiryID int64, update models.ReservationUpdate) error {
	if ownerID < 1 {
		return ErrInvalidInput
	}
	return s.updateReservation(inquiryID, ownerID, "owner", update)
}

func (s *inquiryService) CancelReservation(inquiryID int64) error {
	return s.cancelReservation(inquiryID, 0)
}

func (s *inquiryService) CancelReservationForOwner(ownerID, inquiryID int64) error {
	if ownerID < 1 {
		return ErrInvalidInput
	}
	return s.cancelReservation(inquiryID, ownerID)
}

func (s *inquiryService) cancelReservation(inquiryID, ownerID int64) error {
	if inquiryID < 1 {
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

	query := `SELECT COALESCE(i.commission_due, 0), i.commission_paid_amount FROM inquiries i JOIN cabins c ON c.id = i.cabin_id WHERE i.id = ? AND i.status IN ('pending', 'reserved') AND i.cancelled_at IS NULL`
	args := []any{inquiryID}
	if ownerID > 0 {
		query += ` AND c.owner_id = ?`
		args = append(args, ownerID)
	}
	var commission, paid int64
	err = tx.QueryRow(query, args...).Scan(&commission, &paid)
	if errors.Is(err, sql.ErrNoRows) {
		var exists int
		existsQuery := `SELECT EXISTS (SELECT 1 FROM inquiries i JOIN cabins c ON c.id = i.cabin_id WHERE i.id = ?`
		existsArgs := []any{inquiryID}
		if ownerID > 0 {
			existsQuery += ` AND c.owner_id = ?`
			existsArgs = append(existsArgs, ownerID)
		}
		existsQuery += `)`
		if err := tx.QueryRow(existsQuery, existsArgs...).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			return ErrNotFound
		}
		return ErrInvalidInput
	}
	if err != nil {
		return err
	}
	updateQuery := `UPDATE inquiries SET status = 'lost', cancelled_at = CURRENT_TIMESTAMP, cancelled_from_status = status, commission_due = commission_paid_amount WHERE id = ? AND status IN ('pending', 'reserved') AND cancelled_at IS NULL`
	updateArgs := []any{inquiryID}
	if ownerID > 0 {
		updateQuery += ` AND EXISTS (SELECT 1 FROM cabins c WHERE c.id = inquiries.cabin_id AND c.owner_id = ?)`
		updateArgs = append(updateArgs, ownerID)
	}
	result, err := tx.Exec(updateQuery, updateArgs...)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated != 1 {
		return ErrInvalidInput
	}
	if commission != paid {
		actorRole := "admin"
		if ownerID > 0 {
			actorRole = "owner"
		}
		if _, err := tx.Exec(`INSERT INTO commission_adjustments (inquiry_id, previous_amount, new_amount, amount, actor_role, actor_owner_id) VALUES (?, ?, ?, ?, ?, NULLIF(?, 0))`, inquiryID, commission, paid, paid-commission, actorRole, ownerID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *inquiryService) UpdatePendingInquiry(inquiryID int64, update models.InquiryUpdate) error {
	return s.updatePendingInquiry(inquiryID, 0, update)
}

func (s *inquiryService) UpdatePendingInquiryForOwner(ownerID, inquiryID int64, update models.InquiryUpdate) error {
	if ownerID < 1 {
		return ErrInvalidInput
	}
	return s.updatePendingInquiry(inquiryID, ownerID, update)
}

func (s *inquiryService) updatePendingInquiry(inquiryID, ownerID int64, update models.InquiryUpdate) error {
	if inquiryID < 1 {
		return ErrInvalidInput
	}
	update.GuestName = strings.TrimSpace(update.GuestName)
	update.GuestPhone = helpers.DigitsOnly(update.GuestPhone)
	update.GuestEmail = strings.TrimSpace(update.GuestEmail)
	if update.GuestName == "" || len(update.GuestName) > 120 || len(update.GuestPhone) < 8 || len(update.GuestPhone) > 15 || len(update.GuestEmail) > 254 || update.Guests < 1 {
		return ErrInvalidInput
	}
	address, err := mail.ParseAddress(update.GuestEmail)
	if err != nil || address.Address != update.GuestEmail {
		return ErrInvalidInput
	}
	from, err := time.Parse(time.DateOnly, update.DateFrom)
	if err != nil {
		return ErrInvalidInput
	}
	to, err := time.Parse(time.DateOnly, update.DateTo)
	if err != nil {
		return ErrInvalidInput
	}
	if update.StayType == models.StayDayUse {
		if !helpers.SameDate(from, to) {
			return ErrInvalidInput
		}
	} else if update.StayType != models.StayOvernight || !to.After(from) || to.Sub(from).Hours()/24 > 365 {
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

	query := inquirySelect + ` WHERE i.id = ? AND i.status = 'pending' AND i.cancelled_at IS NULL`
	args := []any{inquiryID}
	if ownerID > 0 {
		query += ` AND c.owner_id = ?`
		args = append(args, ownerID)
	}
	current, err := scanInquiry(tx.QueryRow(query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	cabin, err := scanCabin(tx.QueryRow(cabinSelect+` WHERE c.id = ?`, current.CabinID))
	if err != nil {
		return err
	}
	if update.StayType == models.StayDayUse && !cabin.AcceptsDayUse || update.StayType == models.StayOvernight && !cabin.AcceptsOvernight {
		return ErrStayTypeNotAccepted
	}
	quoteEnd := to
	dateTo := update.DateTo
	if update.StayType == models.StayDayUse {
		quoteEnd = from
		dateTo = from.AddDate(0, 0, 1).Format(time.DateOnly)
	}
	quote, err := Quote(QuoteInput{
		Type: update.StayType, StartDate: from, EndDate: quoteEnd, Guests: update.Guests,
		IncludedGuests: cabin.IncludedGuests, ExtraGuestPrice: cabin.ExtraGuestPrice,
		DayPrice: cabin.DayPrice, NightPrice: cabin.NightPrice,
	})
	if err != nil {
		return err
	}
	estimated, err := AddDepositToQuote(quote, current.DepositCOP)
	if err != nil {
		return err
	}
	startTime, endTime := cabin.CheckIn, cabin.CheckOut
	if update.StayType == models.StayDayUse {
		startTime, endTime = cabin.DayStart, cabin.DayEnd
	}
	startAt, readyAt, err := reservationWindow(cabin, update.StayType, update.DateFrom, update.DateTo)
	if err != nil {
		return err
	}
	var occupied int
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM inquiries i
			WHERE i.id <> ? AND i.cabin_id = ? AND i.status = 'reserved'
			  AND datetime(?) < datetime(
				  (CASE WHEN i.stay_type = 'day_use' THEN i.date_from ELSE i.date_to END) || ' ' || i.end_time,
				  printf('+%d hours', i.cleaning_hours)
			  )
			  AND datetime(?) > datetime(i.date_from || ' ' || i.start_time)
		) OR EXISTS (
			SELECT 1 FROM cabin_blocks b WHERE b.cabin_id = ? AND b.date_from < ? AND b.date_to > ?
		)
	`, inquiryID, cabin.ID, startAt.Format(time.DateTime), readyAt.Format(time.DateTime), cabin.ID, dateTo, update.DateFrom).Scan(&occupied)
	if err != nil {
		return err
	}
	if occupied != 0 {
		return ErrUnavailable
	}
	updateQuery := `
		UPDATE inquiries
		SET guest_name = ?, guest_phone = ?, guest_email = ?, stay_type = ?, date_from = ?, date_to = ?,
		    start_time = ?, end_time = ?, cleaning_hours = ?, guests = ?, estimated_total = ?
		WHERE id = ? AND status = 'pending' AND cancelled_at IS NULL
	`
	updateArgs := []any{update.GuestName, update.GuestPhone, update.GuestEmail, update.StayType, update.DateFrom, dateTo,
		startTime, endTime, cabin.CleaningHours, update.Guests, estimated, inquiryID}
	if ownerID > 0 {
		updateQuery += ` AND EXISTS (SELECT 1 FROM cabins c WHERE c.id = inquiries.cabin_id AND c.owner_id = ?)`
		updateArgs = append(updateArgs, ownerID)
	}
	result, err := tx.Exec(updateQuery, updateArgs...)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated != 1 {
		return ErrNotFound
	}
	return tx.Commit()
}

func (s *inquiryService) updateReservation(inquiryID, ownerID int64, actorRole string, update models.ReservationUpdate) error {
	if inquiryID < 1 {
		return ErrInvalidInput
	}
	update.GuestName = strings.TrimSpace(update.GuestName)
	update.GuestPhone = helpers.DigitsOnly(update.GuestPhone)
	update.GuestEmail = strings.TrimSpace(update.GuestEmail)
	if update.GuestName == "" || len(update.GuestName) > 120 || len(update.GuestPhone) < 8 || len(update.GuestPhone) > 15 || len(update.GuestEmail) > 254 || update.Guests < 1 {
		return ErrInvalidInput
	}
	address, err := mail.ParseAddress(update.GuestEmail)
	if err != nil || address.Address != update.GuestEmail {
		return ErrInvalidInput
	}
	from, err := time.Parse(time.DateOnly, update.DateFrom)
	if err != nil {
		return ErrInvalidInput
	}
	to, err := time.Parse(time.DateOnly, update.DateTo)
	if err != nil {
		return ErrInvalidInput
	}
	if update.StayType == models.StayDayUse {
		if !helpers.SameDate(from, to) {
			return ErrInvalidInput
		}
	} else if update.StayType != models.StayOvernight || !to.After(from) || to.Sub(from).Hours()/24 > 365 {
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

	query := inquirySelect + ` WHERE i.id = ? AND i.status = 'reserved'`
	args := []any{inquiryID}
	if ownerID > 0 {
		query += ` AND c.owner_id = ?`
		args = append(args, ownerID)
	}
	current, err := scanInquiry(tx.QueryRow(query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	cabin, err := scanCabin(tx.QueryRow(cabinSelect+` WHERE c.id = ?`, current.CabinID))
	if err != nil {
		return err
	}
	if update.StayType == models.StayDayUse && !cabin.AcceptsDayUse || update.StayType == models.StayOvernight && !cabin.AcceptsOvernight {
		return ErrStayTypeNotAccepted
	}

	dateTo := update.DateTo
	if update.StayType == models.StayDayUse {
		dateTo = from.AddDate(0, 0, 1).Format(time.DateOnly)
	}
	startTime, endTime, cleaningHours := current.StartTime, current.EndTime, current.CleaningHours
	if update.StayType != current.StayType {
		if update.StayType == models.StayDayUse {
			startTime, endTime = cabin.DayStart, cabin.DayEnd
		} else {
			startTime, endTime = cabin.CheckIn, cabin.CheckOut
		}
		cleaningHours = cabin.CleaningHours
	}

	priceChanged := update.Guests != current.Guests || update.StayType != current.StayType || update.DateFrom != current.DateFrom || dateTo != current.DateTo
	finalTotal := current.FinalTotal
	if update.UseFinalTotal {
		if update.FinalTotal < current.DepositCOP || update.FinalTotal > MaxAmountCOP {
			return ErrInvalidInput
		}
		finalTotal = update.FinalTotal
	} else if priceChanged {
		quoteEnd := to
		if update.StayType == models.StayDayUse {
			quoteEnd = from
		}
		quote, quoteErr := Quote(QuoteInput{
			Type: update.StayType, StartDate: from, EndDate: quoteEnd, Guests: update.Guests,
			IncludedGuests: cabin.IncludedGuests, ExtraGuestPrice: cabin.ExtraGuestPrice,
			DayPrice: cabin.DayPrice, NightPrice: cabin.NightPrice,
		})
		if quoteErr != nil {
			return quoteErr
		}
		finalTotal, err = AddDepositToQuote(quote, current.DepositCOP)
		if err != nil {
			return err
		}
	}
	commission, err := Commission(finalTotal-current.DepositCOP, current.CommissionBP)
	if err != nil {
		return err
	}
	result, err := tx.Exec(`
		UPDATE inquiries
		SET guest_name = ?, guest_phone = ?, guest_email = ?, stay_type = ?, date_from = ?, date_to = ?,
		    start_time = ?, end_time = ?, cleaning_hours = ?, guests = ?, final_total = ?, commission_due = ?
		WHERE id = ? AND status = 'reserved'
	`, update.GuestName, update.GuestPhone, update.GuestEmail, update.StayType, update.DateFrom, dateTo,
		startTime, endTime, cleaningHours, update.Guests, finalTotal, commission, inquiryID)
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
		return ErrNotFound
	}
	if commission != current.Commission {
		_, err = tx.Exec(`
			INSERT INTO commission_adjustments (inquiry_id, previous_amount, new_amount, amount, actor_role, actor_owner_id)
			VALUES (?, ?, ?, ?, ?, NULLIF(?, 0))
		`, inquiryID, current.Commission, commission, commission-current.Commission, actorRole, ownerID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
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

func (s *inquiryService) ReservationsForOwner(ownerID int64) ([]models.Inquiry, error) {
	if ownerID < 1 {
		return nil, ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(inquirySelect+` WHERE c.owner_id = ? AND (i.status IN ('pending', 'reserved') OR i.cancelled_at IS NOT NULL) ORDER BY i.date_from DESC, i.id DESC`, ownerID)
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
	var deposit int64
	err = tx.QueryRow(`
		SELECT i.status, cs.commission_bps, i.deposit_cop
		FROM inquiries i JOIN commission_settings cs ON cs.id = 1
		WHERE i.report_token_hash = ?
	`, hashToken(token)).Scan(&current, &commissionBP, &deposit)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if current != models.InquiryPending {
		return ErrInvalidInput
	}
	if status == models.InquiryReserved && finalTotal < deposit {
		return ErrInvalidInput
	}

	var result sql.Result
	if status == models.InquiryLost {
		result, err = tx.Exec(`UPDATE inquiries SET status = 'lost', resolved_at = CURRENT_TIMESTAMP WHERE report_token_hash = ? AND status = 'pending'`, hashToken(token))
	} else {
		commission, err := Commission(finalTotal-deposit, commissionBP)
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
				  AND datetime(inquiries.date_from || ' ' || inquiries.start_time) < datetime(
					  (CASE WHEN other.stay_type = 'day_use' THEN other.date_from ELSE other.date_to END) || ' ' || other.end_time,
					  printf('+%d hours', other.cleaning_hours)
				  )
				  AND datetime(
					  (CASE WHEN inquiries.stay_type = 'day_use' THEN inquiries.date_from ELSE inquiries.date_to END) || ' ' || inquiries.end_time,
					  printf('+%d hours', inquiries.cleaning_hours)
				  ) > datetime(other.date_from || ' ' || other.start_time)
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
	       i.stay_type, i.date_from, i.date_to, i.start_time, i.end_time, i.cleaning_hours, i.guests, i.estimated_total, i.deposit_cop,
	       CASE WHEN i.cancelled_at IS NOT NULL THEN 'cancelled' ELSE i.status END, COALESCE(i.cancelled_from_status, ''),
	       COALESCE(i.final_total, 0), COALESCE(i.commission_bps, cs.commission_bps), COALESCE(i.commission_due, 0), COALESCE(i.commission_paid_at, ''),
	       i.commission_paid_amount, COALESCE(i.commission_due, 0) - i.commission_paid_amount, i.created_at
	FROM inquiries i JOIN cabins c ON c.id = i.cabin_id LEFT JOIN owners o ON o.id = c.owner_id JOIN commission_settings cs ON cs.id = 1
`

func scanInquiry(row rowScanner) (models.Inquiry, error) {
	var inquiry models.Inquiry
	err := row.Scan(
		&inquiry.ID, &inquiry.CabinID, &inquiry.CabinName, &inquiry.OwnerName, &inquiry.WhatsApp,
		&inquiry.GuestName, &inquiry.GuestPhone, &inquiry.GuestEmail, &inquiry.StayType, &inquiry.DateFrom, &inquiry.DateTo, &inquiry.StartTime, &inquiry.EndTime, &inquiry.CleaningHours,
		&inquiry.Guests, &inquiry.Estimated, &inquiry.DepositCOP, &inquiry.Status, &inquiry.CancelledFromStatus, &inquiry.FinalTotal, &inquiry.CommissionBP, &inquiry.Commission, &inquiry.CommissionPaidAt,
		&inquiry.CommissionPaidAmount, &inquiry.CommissionBalance, &inquiry.CreatedAt,
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

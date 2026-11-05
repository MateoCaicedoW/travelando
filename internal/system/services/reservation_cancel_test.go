package services

import (
	"errors"
	"testing"

	"travelando/internal/system/models"
)

func TestCancelReservationReleasesDatesAndWaivesUnpaidCommission(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño', '573001234567', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price)
		VALUES (1, 'Cabaña', 2, 0, 100000, 200000);
		UPDATE commission_settings SET commission_bps = 1000 WHERE id = 1;
	`); err != nil {
		t.Fatal(err)
	}

	inquiries := NewInquiryService(conn)
	request := models.Inquiry{
		CabinID: 1, GuestName: "Huésped", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: "guest@example.com", TermsAccepted: true,
		StayType: models.StayDayUse, DateFrom: "2099-02-10", DateTo: "2099-02-10", Guests: 2,
	}
	token, err := inquiries.CreateInquiry(&request)
	if err != nil {
		t.Fatal(err)
	}
	if err := inquiries.ResolveInquiry(token, models.InquiryReserved, 100000); err != nil {
		t.Fatal(err)
	}
	reservation, err := inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}

	if err := inquiries.CancelReservation(reservation.ID); err != nil {
		t.Fatal(err)
	}
	cancelled, err := inquiries.AdminReservation(reservation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != models.InquiryCancelled || cancelled.CommissionBalance != 0 || cancelled.CommissionPaidAmount != 0 {
		t.Fatalf("cancelled reservation = %+v, want cancelled with no commission balance", cancelled)
	}
	if len(cancelled.CommissionAdjustments) != 1 || cancelled.CommissionAdjustments[0].Amount != -10000 {
		t.Fatalf("commission adjustments = %+v, want -10000 waiver", cancelled.CommissionAdjustments)
	}

	if err := inquiries.CancelReservation(reservation.ID); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("repeated cancellation error = %v, want ErrInvalidInput", err)
	}

	rebook := request
	rebook.GuestEmail = "new@example.com"
	rebook.GuestName = "Nuevo huésped"
	newToken, err := inquiries.CreateInquiry(&rebook)
	if err != nil {
		t.Fatalf("create inquiry on released dates: %v", err)
	}
	if err := inquiries.ResolveInquiry(newToken, models.InquiryReserved, 100000); err != nil {
		t.Fatalf("confirm reservation on released dates: %v", err)
	}
}

func TestCancelReservationKeepsPaidCommissionInOwnerHistory(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño', '573001234567', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price)
		VALUES (1, 'Cabaña', 2, 0, 100000, 200000);
		UPDATE commission_settings SET commission_bps = 1000 WHERE id = 1;
	`); err != nil {
		t.Fatal(err)
	}
	inquiries := NewInquiryService(conn)
	request := models.Inquiry{
		CabinID: 1, GuestName: "Huésped", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: "guest@example.com", TermsAccepted: true,
		StayType: models.StayDayUse, DateFrom: "2099-03-10", DateTo: "2099-03-10", Guests: 2,
	}
	token, err := inquiries.CreateInquiry(&request)
	if err != nil {
		t.Fatal(err)
	}
	if err := inquiries.ResolveInquiry(token, models.InquiryReserved, 100000); err != nil {
		t.Fatal(err)
	}
	reservation, err := inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if err := inquiries.MarkCommissionPaid(reservation.ID); err != nil {
		t.Fatal(err)
	}
	if err := inquiries.CancelReservation(reservation.ID); err != nil {
		t.Fatal(err)
	}
	cancelled, err := inquiries.AdminReservation(reservation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.CommissionPaidAmount != 10000 || cancelled.CommissionBalance != 0 {
		t.Fatalf("cancelled commission = paid %d, balance %d; want paid 10000, balance 0", cancelled.CommissionPaidAmount, cancelled.CommissionBalance)
	}
	owners, err := NewOwnerService(conn).Owners()
	if err != nil {
		t.Fatal(err)
	}
	if len(owners) != 1 || owners[0].CommissionPaid != 10000 || owners[0].CommissionDue != 0 || owners[0].CommissionCredit != 0 {
		t.Fatalf("owner commission summary = %+v, want paid 10000 and no balance", owners)
	}
}

func TestUpdatePendingInquiryRequotesAndKeepsPendingStatus(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño', '573001234567', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price)
		VALUES (1, 'Cabaña', 2, 15000, 100000, 200000);
		UPDATE commission_settings SET deposit_cop = 5000 WHERE id = 1;
	`); err != nil {
		t.Fatal(err)
	}
	inquiries := NewInquiryService(conn)
	request := models.Inquiry{
		CabinID: 1, GuestName: "Huésped", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: "guest@example.com", TermsAccepted: true,
		StayType: models.StayDayUse, DateFrom: "2099-02-10", DateTo: "2099-02-10", Guests: 2,
	}
	token, err := inquiries.CreateInquiry(&request)
	if err != nil {
		t.Fatal(err)
	}
	inquiry, err := inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if err := inquiries.UpdatePendingInquiry(inquiry.ID, models.InquiryUpdate{
		GuestName: "Huésped editado", GuestPhone: "573009876543", GuestEmail: "edited@example.com",
		StayType: models.StayOvernight, DateFrom: "2099-02-12", DateTo: "2099-02-14", Guests: 4,
	}); err != nil {
		t.Fatal(err)
	}
	updated, err := inquiries.AdminReservation(inquiry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != models.InquiryPending || updated.GuestName != "Huésped editado" || updated.DateFrom != "2099-02-12" || updated.DateTo != "2099-02-14" || updated.Guests != 4 || updated.Estimated != 435000 || updated.FinalTotal != 0 {
		t.Fatalf("updated pending inquiry = %+v, want edited details, 435000 estimate, and pending status", updated)
	}
}

func TestUpdatePendingInquiryRejectsDatesOccupiedByConfirmedReservation(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño', '573001234567', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price)
		VALUES (1, 'Cabaña', 2, 0, 100000, 200000);
	`); err != nil {
		t.Fatal(err)
	}
	inquiries := NewInquiryService(conn)
	makeInquiry := func(name, email, date string) int64 {
		t.Helper()
		request := models.Inquiry{
			CabinID: 1, GuestName: name, GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: email, TermsAccepted: true,
			StayType: models.StayDayUse, DateFrom: date, DateTo: date, Guests: 2,
		}
		token, err := inquiries.CreateInquiry(&request)
		if err != nil {
			t.Fatal(err)
		}
		inquiry, err := inquiries.InquiryByToken(token)
		if err != nil {
			t.Fatal(err)
		}
		return inquiry.ID
	}
	pendingID := makeInquiry("Pendiente", "pending@example.com", "2099-04-10")
	reservedToken, err := inquiries.CreateInquiry(&models.Inquiry{
		CabinID: 1, GuestName: "Confirmada", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: "reserved@example.com", TermsAccepted: true,
		StayType: models.StayDayUse, DateFrom: "2099-04-11", DateTo: "2099-04-11", Guests: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := inquiries.ResolveInquiry(reservedToken, models.InquiryReserved, 100000); err != nil {
		t.Fatal(err)
	}
	err = inquiries.UpdatePendingInquiry(pendingID, models.InquiryUpdate{
		GuestName: "Pendiente", GuestPhone: "573001234567", GuestEmail: "pending@example.com",
		StayType: models.StayDayUse, DateFrom: "2099-04-11", DateTo: "2099-04-11", Guests: 2,
	})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("pending edit overlapping confirmed dates = %v, want ErrUnavailable", err)
	}
}

func TestCancelPendingInquiryIsRecordedAsCancelled(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño', '573001234567', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price)
		VALUES (1, 'Cabaña', 2, 0, 100000, 200000);
	`); err != nil {
		t.Fatal(err)
	}
	inquiries := NewInquiryService(conn)
	request := models.Inquiry{
		CabinID: 1, GuestName: "Huésped", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: "guest@example.com", TermsAccepted: true,
		StayType: models.StayDayUse, DateFrom: "2099-05-10", DateTo: "2099-05-10", Guests: 2,
	}
	token, err := inquiries.CreateInquiry(&request)
	if err != nil {
		t.Fatal(err)
	}
	inquiry, err := inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if err := inquiries.CancelReservation(inquiry.ID); err != nil {
		t.Fatal(err)
	}
	cancelled, err := inquiries.AdminReservation(inquiry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != models.InquiryCancelled || cancelled.CancelledFromStatus != models.InquiryPending {
		t.Fatalf("cancelled pending inquiry = %+v, want cancellation from pending", cancelled)
	}
}

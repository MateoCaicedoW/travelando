package services

import (
	"errors"
	"testing"

	"travelando/internal/system/models"
)

func TestUpdateReservedInquiryRequotesAndKeepsCommissionRateAfterPayment(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño', '573001234567', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price, day_start, day_end, check_in, check_out, cleaning_hours)
		VALUES (1, 'Cabaña', 2, 15000, 100000, 200000, '08:00', '18:00', '15:00', '12:00', 3);
		UPDATE commission_settings SET commission_bps = 1000 WHERE id = 1;
	`); err != nil {
		t.Fatal(err)
	}

	inquiries := NewInquiryService(conn)
	request := models.Inquiry{
		CabinID: 1, GuestName: "Huésped", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: "guest@example.com", TermsAccepted: true,
		StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-10", Guests: 2,
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
	if reservation.CommissionBP != 1000 || reservation.Commission != 10000 {
		t.Fatalf("initial commission = %d at %d bps, want 10000 at 1000 bps", reservation.Commission, reservation.CommissionBP)
	}
	if err := inquiries.MarkCommissionPaid(reservation.ID); err != nil {
		t.Fatal(err)
	}
	if err := NewCommissionService(conn).UpdateCommissionRate(2000); err != nil {
		t.Fatal(err)
	}

	err = inquiries.UpdateReservation(reservation.ID, models.ReservationUpdate{
		GuestName: "Huésped actualizado", GuestPhone: "573009876543", GuestEmail: "new@example.com",
		StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-10", Guests: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Guests != 5 || updated.FinalTotal != 145000 {
		t.Fatalf("updated guests/total = %d/%d, want 5/145000", updated.Guests, updated.FinalTotal)
	}
	if updated.CommissionBP != 1000 || updated.Commission != 14500 {
		t.Fatalf("updated commission = %d at %d bps, want 14500 at original 1000 bps", updated.Commission, updated.CommissionBP)
	}
	if updated.CommissionPaidAmount != 10000 || updated.CommissionBalance != 4500 {
		t.Fatalf("paid/balance = %d/%d, want 10000/4500", updated.CommissionPaidAmount, updated.CommissionBalance)
	}
	var adjustment int64
	if err := database.QueryRow(`SELECT amount FROM commission_adjustments WHERE inquiry_id = ?`, reservation.ID).Scan(&adjustment); err != nil {
		t.Fatal(err)
	}
	if adjustment != 4500 {
		t.Fatalf("commission adjustment = %d, want 4500", adjustment)
	}
	if err := inquiries.MarkCommissionPaid(reservation.ID); err != nil {
		t.Fatal(err)
	}
	updated, err = inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if updated.CommissionPaidAmount != 14500 || updated.CommissionBalance != 0 {
		t.Fatalf("paid/balance after settling adjustment = %d/%d, want 14500/0", updated.CommissionPaidAmount, updated.CommissionBalance)
	}
	if err := inquiries.UpdateReservation(reservation.ID, models.ReservationUpdate{
		GuestName: "Huésped actualizado", GuestPhone: "573009876543", GuestEmail: "new@example.com",
		StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-10", Guests: 2,
	}); err != nil {
		t.Fatal(err)
	}
	updated, err = inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if updated.FinalTotal != 100000 || updated.Commission != 10000 || updated.CommissionBalance != -4500 {
		t.Fatalf("reduced commission = total %d, commission %d, balance %d; want 100000, 10000, -4500", updated.FinalTotal, updated.Commission, updated.CommissionBalance)
	}
	var adjustments int
	if err := database.QueryRow(`SELECT COUNT(*) FROM commission_adjustments WHERE inquiry_id = ?`, reservation.ID).Scan(&adjustments); err != nil {
		t.Fatal(err)
	}
	if adjustments != 2 {
		t.Fatalf("adjustment count = %d, want increase and decrease records", adjustments)
	}
	if err := inquiries.UpdateReservation(reservation.ID, models.ReservationUpdate{
		GuestName: "Huésped actualizado", GuestPhone: "573009876543", GuestEmail: "new@example.com",
		StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-10", Guests: 2,
		FinalTotal: 120000, UseFinalTotal: true,
	}); err != nil {
		t.Fatal(err)
	}
	updated, err = inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if updated.FinalTotal != 120000 || updated.Commission != 12000 || updated.CommissionBalance != -2500 {
		t.Fatalf("manual total commission = %d/%d/%d, want 120000/12000/-2500", updated.FinalTotal, updated.Commission, updated.CommissionBalance)
	}
	owners, err := NewOwnerService(conn).Owners()
	if err != nil {
		t.Fatal(err)
	}
	if len(owners) != 1 || owners[0].CommissionPaid != 14500 || owners[0].CommissionDue != 0 || owners[0].CommissionCredit != 2500 {
		t.Fatalf("owner commission totals = %+v, want paid 14500, due 0, credit 2500", owners)
	}
	withHistory, err := inquiries.Reservation(reservation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(withHistory.CommissionAdjustments) != 3 || withHistory.CommissionAdjustments[0].ActorRole != "admin" {
		t.Fatalf("reservation adjustment history = %+v, want 3 admin changes", withHistory.CommissionAdjustments)
	}
}

func TestOwnerCannotUpdateAnotherOwnersReservation(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño', '573001234567', ''), ('Otro dueño', '573009876543', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price)
		VALUES (1, 'Cabaña', 2, 10000, 100000, 200000);
		UPDATE commission_settings SET commission_bps = 1000 WHERE id = 1;
	`); err != nil {
		t.Fatal(err)
	}
	inquiries := NewInquiryService(conn)
	request := models.Inquiry{
		CabinID: 1, GuestName: "Huésped", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: "guest@example.com", TermsAccepted: true,
		StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-10", Guests: 2,
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
	err = inquiries.UpdateReservationForOwner(2, reservation.ID, models.ReservationUpdate{
		GuestName: "Alterado", GuestPhone: "573001234567", GuestEmail: "guest@example.com",
		StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-10", Guests: 2,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("owner update error = %v, want ErrNotFound", err)
	}
	if err := inquiries.UpdateReservationForOwner(1, reservation.ID, models.ReservationUpdate{
		GuestName: "Huésped", GuestPhone: "573001234567", GuestEmail: "guest@example.com",
		StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-10", Guests: 3,
	}); err != nil {
		t.Fatal(err)
	}
	updated, err := inquiries.ReservationForOwner(1, reservation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Guests != 3 || updated.FinalTotal != 110000 || updated.Commission != 11000 || len(updated.CommissionAdjustments) != 1 || updated.CommissionAdjustments[0].ActorRole != "owner" {
		t.Fatalf("owner update result = %+v, want recalculation and owner adjustment", updated)
	}
}

func TestReservationsForOwnerScopesConfirmedReservations(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño uno', '573001234567', ''), ('Dueño dos', '573009876543', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price)
		VALUES (1, 'Cabaña uno', 2, 0, 100000, 200000), (2, 'Cabaña dos', 2, 0, 100000, 200000);
	`); err != nil {
		t.Fatal(err)
	}
	inquiries := NewInquiryService(conn)
	create := func(cabinID int64, email, date string, confirm bool) string {
		t.Helper()
		request := models.Inquiry{
			CabinID: cabinID, GuestName: "Huésped", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: email, TermsAccepted: true,
			StayType: models.StayDayUse, DateFrom: date, DateTo: date, Guests: 2,
		}
		token, err := inquiries.CreateInquiry(&request)
		if err != nil {
			t.Fatal(err)
		}
		if confirm {
			if err := inquiries.ResolveInquiry(token, models.InquiryReserved, 100000); err != nil {
				t.Fatal(err)
			}
		}
		return token
	}
	create(1, "owner-one@example.com", "2099-08-10", true)
	create(1, "pending@example.com", "2099-08-11", false)
	create(2, "owner-two@example.com", "2099-08-12", true)
	cancelledToken := create(1, "cancelled@example.com", "2099-08-13", true)
	cancelledInquiry, err := inquiries.InquiryByToken(cancelledToken)
	if err != nil {
		t.Fatal(err)
	}
	if err := inquiries.CancelReservation(cancelledInquiry.ID); err != nil {
		t.Fatal(err)
	}

	ownerOne, err := inquiries.ReservationsForOwner(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(ownerOne) != 3 || ownerOne[0].CabinID != 1 || ownerOne[0].Status != models.InquiryCancelled || ownerOne[1].CabinID != 1 || ownerOne[1].Status != models.InquiryPending || ownerOne[2].CabinID != 1 || ownerOne[2].Status != models.InquiryReserved {
		t.Fatalf("owner one reservations = %+v, want cancelled, pending, and confirmed records for cabin one", ownerOne)
	}
	ownerTwo, err := inquiries.ReservationsForOwner(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(ownerTwo) != 1 || ownerTwo[0].CabinID != 2 {
		t.Fatalf("owner two reservations = %+v, want only cabin two reservation", ownerTwo)
	}
}

func TestOwnerCanCancelOnlyTheirOwnReservation(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño uno', '573001234567', ''), ('Dueño dos', '573009876543', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price)
		VALUES (1, 'Cabaña uno', 2, 0, 100000, 200000), (2, 'Cabaña dos', 2, 0, 100000, 200000);
		UPDATE commission_settings SET commission_bps = 1000 WHERE id = 1;
	`); err != nil {
		t.Fatal(err)
	}
	inquiries := NewInquiryService(conn)
	token, err := inquiries.CreateInquiry(&models.Inquiry{
		CabinID: 1, GuestName: "Huésped", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: "guest@example.com", TermsAccepted: true,
		StayType: models.StayDayUse, DateFrom: "2099-10-10", DateTo: "2099-10-10", Guests: 2,
	})
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
	if err := inquiries.CancelReservationForOwner(2, reservation.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("other owner cancellation error = %v, want ErrNotFound", err)
	}
	if err := inquiries.CancelReservationForOwner(1, reservation.ID); err != nil {
		t.Fatal(err)
	}
	cancelled, err := inquiries.InquiryForOwner(1, reservation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != models.InquiryCancelled || cancelled.CancelledFromStatus != models.InquiryReserved || len(cancelled.CommissionAdjustments) != 1 || cancelled.CommissionAdjustments[0].ActorRole != "owner" {
		t.Fatalf("owner-cancelled reservation = %+v, want cancelled history with owner commission adjustment", cancelled)
	}
}

func TestOwnerCanUpdateOnlyTheirOwnPendingInquiry(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño uno', '573001234567', ''), ('Dueño dos', '573009876543', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price)
		VALUES (1, 'Cabaña uno', 2, 15000, 100000, 200000), (2, 'Cabaña dos', 2, 0, 100000, 200000);
	`); err != nil {
		t.Fatal(err)
	}
	inquiries := NewInquiryService(conn)
	request := models.Inquiry{
		CabinID: 1, GuestName: "Huésped", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: "guest@example.com", TermsAccepted: true,
		StayType: models.StayDayUse, DateFrom: "2099-09-10", DateTo: "2099-09-10", Guests: 2,
	}
	token, err := inquiries.CreateInquiry(&request)
	if err != nil {
		t.Fatal(err)
	}
	inquiry, err := inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}
	update := models.InquiryUpdate{
		GuestName: "Huésped editado", GuestPhone: "573001234567", GuestEmail: "guest@example.com",
		StayType: models.StayDayUse, DateFrom: "2099-09-10", DateTo: "2099-09-10", Guests: 3,
	}
	if err := inquiries.UpdatePendingInquiryForOwner(2, inquiry.ID, update); !errors.Is(err, ErrNotFound) {
		t.Fatalf("other owner pending edit error = %v, want ErrNotFound", err)
	}
	if err := inquiries.UpdatePendingInquiryForOwner(1, inquiry.ID, update); err != nil {
		t.Fatal(err)
	}
	updated, err := inquiries.InquiryForOwner(1, inquiry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != models.InquiryPending || updated.GuestName != "Huésped editado" || updated.Estimated != 115000 {
		t.Fatalf("updated owner pending inquiry = %+v, want pending and estimated total 115000", updated)
	}
}

func TestUpdateReservationRejectsOverlappingDates(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño', '573001234567', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price, day_start, day_end, check_in, check_out, cleaning_hours)
		VALUES (1, 'Cabaña', 2, 0, 100000, 200000, '08:00', '18:00', '15:00', '12:00', 3);
	`); err != nil {
		t.Fatal(err)
	}
	inquiries := NewInquiryService(conn)
	makeReservation := func(date string) (*models.Inquiry, string) {
		t.Helper()
		request := models.Inquiry{
			CabinID: 1, GuestName: "Huésped", GuestCedula: "12345678", GuestPhone: "573001234567", GuestEmail: "guest@example.com", TermsAccepted: true,
			StayType: models.StayDayUse, DateFrom: date, DateTo: date, Guests: 2,
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
		return reservation, token
	}
	first, _ := makeReservation("2027-02-10")
	second, _ := makeReservation("2027-02-11")
	err := inquiries.UpdateReservation(second.ID, models.ReservationUpdate{
		GuestName: "Huésped", GuestPhone: "573001234567", GuestEmail: "guest@example.com",
		StayType: models.StayDayUse, DateFrom: "2027-02-10", DateTo: "2027-02-10", Guests: 2,
	})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("overlapping edit error = %v, want ErrUnavailable (existing reservation %d)", err, first.ID)
	}
}

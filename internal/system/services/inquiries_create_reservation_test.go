package services

import (
	"errors"
	"testing"

	"travelando/internal/system/models"
)

func TestCreateReservationSetsReservedAndCommission(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash, active) VALUES ('Dueño', '573001234567', '', 1);
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price, day_start, day_end, check_in, check_out, cleaning_hours, active, accepts_day_use, accepts_overnight)
		VALUES (1, 'Cabaña', 2, 15000, 100000, 200000, '08:00', '18:00', '15:00', '12:00', 3, 1, 1, 1);
		UPDATE commission_settings SET commission_bps = 1000 WHERE id = 1;
	`); err != nil {
		t.Fatal(err)
	}

	inquiries := NewInquiryService(conn)
	id, err := inquiries.CreateReservation(1, &models.ReservationCreate{
		CabinID:     1,
		GuestName:   "María García",
		GuestCedula: "12345678",
		GuestPhone:  "573001234567",
		GuestEmail:  "maria@example.com",
		StayType:    models.StayOvernight,
		DateFrom:    "2027-02-10",
		DateTo:      "2027-02-12",
		Guests:      3,
		DepositCOP:  20000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if id < 1 {
		t.Fatal("reservation id should be set")
	}

	reservation, err := inquiries.AdminReservation(id)
	if err != nil {
		t.Fatal(err)
	}
	if reservation.Status != models.InquiryReserved {
		t.Fatalf("status = %s, want reserved", reservation.Status)
	}
	// 2 noches a 200000 = 400000 + 1 extra guest (15000) = 415000 + deposit 20000 = 435000
	if reservation.FinalTotal != 435000 {
		t.Fatalf("final_total = %d, want 435000", reservation.FinalTotal)
	}
	if reservation.DepositCOP != 20000 {
		t.Fatalf("deposit = %d, want 20000", reservation.DepositCOP)
	}
	if reservation.CommissionBP != 1000 {
		t.Fatalf("commission_bps = %d, want 1000", reservation.CommissionBP)
	}
	// commission on (435000 - 20000) at 1000 bps = 4150
	if reservation.Commission != 41500 {
		t.Fatalf("commission = %d, want 41500", reservation.Commission)
	}
	if reservation.GuestName != "María García" || reservation.GuestEmail != "maria@example.com" {
		t.Fatalf("guest data not saved: %s / %s", reservation.GuestName, reservation.GuestEmail)
	}
}

func TestCreateReservationRejectsOverlappingDates(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash, active) VALUES ('Dueño', '573001234567', '', 1);
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price, day_start, day_end, check_in, check_out, cleaning_hours, active, accepts_day_use, accepts_overnight)
		VALUES (1, 'Cabaña', 2, 15000, 100000, 200000, '08:00', '18:00', '15:00', '12:00', 3, 1, 1, 1);
		UPDATE commission_settings SET commission_bps = 1000 WHERE id = 1;
	`); err != nil {
		t.Fatal(err)
	}

	inquiries := NewInquiryService(conn)
	_, err := inquiries.CreateReservation(1, &models.ReservationCreate{
		CabinID:     1,
		GuestName:   "Huésped 1",
		GuestCedula: "11111111",
		GuestPhone:  "573001111111",
		GuestEmail:  "h1@example.com",
		StayType:    models.StayOvernight,
		DateFrom:    "2027-03-01",
		DateTo:      "2027-03-05",
		Guests:      2,
		DepositCOP:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = inquiries.CreateReservation(1, &models.ReservationCreate{
		CabinID:     1,
		GuestName:   "Huésped 2",
		GuestCedula: "22222222",
		GuestPhone:  "573002222222",
		GuestEmail:  "h2@example.com",
		StayType:    models.StayOvernight,
		DateFrom:    "2027-03-04",
		DateTo:      "2027-03-06",
		Guests:      2,
		DepositCOP:  0,
	})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
}

func TestCreateReservationForOwnerOnlyOwnsCabins(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash, active) VALUES ('Dueño A', '573001234567', '', 1), ('Dueño B', '573009876543', '', 1);
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price, day_start, day_end, check_in, check_out, cleaning_hours, active, accepts_day_use, accepts_overnight)
		VALUES (1, 'Cabaña A', 2, 15000, 100000, 200000, '08:00', '18:00', '15:00', '12:00', 3, 1, 1, 1),
		       (2, 'Cabaña B', 2, 15000, 100000, 200000, '08:00', '18:00', '15:00', '12:00', 3, 1, 1, 1);
		UPDATE commission_settings SET commission_bps = 1000 WHERE id = 1;
	`); err != nil {
		t.Fatal(err)
	}

	inquiries := NewInquiryService(conn)

	// Dueño A crea en su propia cabina (id 1) -> ok
	_, err := inquiries.CreateReservationForOwner(1, &models.ReservationCreate{
		CabinID:     1,
		GuestName:   "Huésped A",
		GuestCedula: "11111111",
		GuestPhone:  "573001111111",
		GuestEmail:  "a@example.com",
		StayType:    models.StayDayUse,
		DateFrom:    "2027-04-01",
		DateTo:      "2027-04-01",
		Guests:      2,
		DepositCOP:  0,
	})
	if err != nil {
		t.Fatalf("owner creating in own cabin failed: %v", err)
	}

	// Dueño A crea en cabina de dueño B (id 2) -> ErrNotFound
	_, err = inquiries.CreateReservationForOwner(1, &models.ReservationCreate{
		CabinID:     2,
		GuestName:   "Huésped A",
		GuestCedula: "11111111",
		GuestPhone:  "573001111111",
		GuestEmail:  "a@example.com",
		StayType:    models.StayDayUse,
		DateFrom:    "2027-04-02",
		DateTo:      "2027-04-02",
		Guests:      2,
		DepositCOP:  0,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

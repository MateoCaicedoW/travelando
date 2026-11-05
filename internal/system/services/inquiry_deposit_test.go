package services

import (
	"testing"

	"travelando/internal/system/models"
)

func TestInquiryEstimateIncludesDepositAndCommissionExcludesIt(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash, cedula) VALUES ('Dueño', '573001234567', 'hash', '123456');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price)
		VALUES (1, 'Cabaña', 2, 25000, 100000, 200000);
		UPDATE commission_settings SET commission_bps = 1000, deposit_cop = 50000 WHERE id = 1;
	`); err != nil {
		t.Fatal(err)
	}

	inquiry := models.Inquiry{
		CabinID: 1, GuestName: "Huésped", GuestPhone: "573001234567", GuestEmail: "guest@example.com",
		StayType: models.StayDayUse, DateFrom: "2026-11-15", DateTo: "2026-11-15", Guests: 2,
	}
	token, err := NewInquiryService(conn).CreateInquiry(&inquiry)
	if err != nil {
		t.Fatal(err)
	}
	if inquiry.DepositCOP != 50000 || inquiry.Estimated != 150000 {
		t.Fatalf("inquiry estimate = %d, deposit = %d; want 150000 total including 50000 deposit", inquiry.Estimated, inquiry.DepositCOP)
	}

	if _, err := database.Exec(`UPDATE commission_settings SET deposit_cop = 100000 WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	inquiries := NewInquiryService(conn)
	if err := inquiries.ResolveInquiry(token, models.InquiryReserved, inquiry.Estimated); err != nil {
		t.Fatal(err)
	}
	confirmed, err := inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if confirmed.DepositCOP != 50000 || confirmed.Commission != 10000 {
		t.Fatalf("confirmed deposit = %d, commission = %d; want snapshot 50000 and lodging-only commission 10000", confirmed.DepositCOP, confirmed.Commission)
	}

	if err := inquiries.UpdateReservation(confirmed.ID, models.ReservationUpdate{
		GuestName: confirmed.GuestName, GuestPhone: confirmed.GuestPhone, GuestEmail: confirmed.GuestEmail,
		StayType: confirmed.StayType, DateFrom: confirmed.DateFrom, DateTo: confirmed.DateFrom,
		Guests: 3,
	}); err != nil {
		t.Fatal(err)
	}
	updated, err := inquiries.InquiryByToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if updated.DepositCOP != 50000 || updated.FinalTotal != 175000 || updated.Commission != 12500 {
		t.Fatalf("updated total=%d deposit=%d commission=%d; want 175000, 50000, 12500", updated.FinalTotal, updated.DepositCOP, updated.Commission)
	}
	if err := inquiries.UpdateReservation(updated.ID, models.ReservationUpdate{
		GuestName: updated.GuestName, GuestPhone: updated.GuestPhone, GuestEmail: updated.GuestEmail,
		StayType: updated.StayType, DateFrom: updated.DateFrom, DateTo: updated.DateFrom,
		Guests: 3, FinalTotal: 49999, UseFinalTotal: true,
	}); err == nil {
		t.Fatal("reservation total below the deposit should be rejected")
	}
}

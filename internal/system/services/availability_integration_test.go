package services

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.leapkit.dev/core/db"
	"travelando/internal/migrations"
	"travelando/internal/system/models"
)

func TestAvailabilityAndConfirmationRespectCleaningWindow(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash) VALUES ('Dueño', '573001234567', '');
		INSERT INTO cabins (owner_id, name, included_guests, extra_guest_price, day_price, night_price, day_start, day_end, check_in, check_out, cleaning_hours)
		VALUES (1, 'Cabaña', 2, 0, 100000, 200000, '08:00', '18:00', '21:00', '12:00', 3);
	`); err != nil {
		t.Fatal(err)
	}

	inquiries := NewInquiryService(conn)
	dayUse := models.Inquiry{
		CabinID: 1, GuestName: "Huésped día", GuestPhone: "573001234567", GuestEmail: "dia@example.com",
		StayType: models.StayDayUse, DateFrom: "2026-11-15", DateTo: "2026-11-15", Guests: 2,
	}
	token, err := inquiries.CreateInquiry(&dayUse)
	if err != nil {
		t.Fatal(err)
	}
	if err := inquiries.ResolveInquiry(token, models.InquiryReserved, 100000); err != nil {
		t.Fatal(err)
	}

	availability := NewAvailabilityService(conn)
	from, _ := time.Parse(time.DateOnly, "2026-11-15")
	to, _ := time.Parse(time.DateOnly, "2026-11-16")
	input := SearchInput{StayType: models.StayOvernight, DateFrom: from, DateTo: to}
	result, err := availability.Search(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 {
		t.Fatalf("overnight starting at cleaning completion should be available, got %d cabins", result.Total)
	}

	if _, err := database.Exec(`UPDATE cabins SET check_in = '20:59' WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	result, err = availability.Search(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 0 {
		t.Fatalf("overnight starting before cleaning completion should be unavailable, got %d cabins", result.Total)
	}
	if _, err := database.Exec(`
		INSERT INTO inquiries (cabin_id, guest_name, guest_phone, guest_email, stay_type, date_from, date_to, start_time, end_time, cleaning_hours, guests, estimated_total, report_token_hash)
		VALUES (1, 'Bypass', '573001234567', 'bypass@example.com', 'overnight', '2026-11-15', '2026-11-16', '20:59', '12:00', 3, 2, 100000, 'trigger-test');
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE inquiries SET status = 'reserved' WHERE report_token_hash = 'trigger-test'`); err == nil {
		t.Fatal("database trigger allowed reservation to bypass cleaning window")
	}

	tooEarly := models.Inquiry{
		CabinID: 1, GuestName: "Huésped noche", GuestPhone: "573001234567", GuestEmail: "noche@example.com",
		StayType: models.StayOvernight, DateFrom: "2026-11-15", DateTo: "2026-11-16", Guests: 2,
	}
	if _, err := inquiries.CreateInquiry(&tooEarly); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("request before cleaning completion error = %v, want ErrUnavailable", err)
	}

	if _, err := database.Exec(`UPDATE cabins SET check_in = '21:00' WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	valid := tooEarly
	token, err = inquiries.CreateInquiry(&valid)
	if err != nil {
		t.Fatal(err)
	}
	if err := inquiries.ResolveInquiry(token, models.InquiryReserved, 200000); err != nil {
		t.Fatalf("reservation at cleaning completion should confirm: %v", err)
	}

	calendar, err := NewCalendarService(conn).Calendar(2026, time.November)
	if err != nil {
		t.Fatal(err)
	}
	if len(calendar.Entries) != 2 {
		t.Fatalf("calendar has %d entries, want day-use and overnight reservations", len(calendar.Entries))
	}
	for _, entry := range calendar.Entries {
		if entry.StayType == models.StayDayUse && entry.ReadyAt != "2026-11-15 21:00:00" {
			t.Errorf("day-use ready at %q, want 2026-11-15 21:00:00", entry.ReadyAt)
		}
	}

	if _, err := database.Exec(`UPDATE cabins SET day_end = '23:00' WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	monthEndUse := models.Inquiry{
		CabinID: 1, GuestName: "Huésped fin de mes", GuestPhone: "573001234567", GuestEmail: "mes@example.com",
		StayType: models.StayDayUse, DateFrom: "2026-12-31", DateTo: "2026-12-31", Guests: 2,
	}
	token, err = inquiries.CreateInquiry(&monthEndUse)
	if err != nil {
		t.Fatal(err)
	}
	if err := inquiries.ResolveInquiry(token, models.InquiryReserved, 100000); err != nil {
		t.Fatal(err)
	}
	calendar, err = NewCalendarService(conn).Calendar(2027, time.January)
	if err != nil {
		t.Fatal(err)
	}
	if len(calendar.Entries) != 1 {
		t.Fatalf("January calendar has %d entries, want cleanup from December reservation", len(calendar.Entries))
	}
	startAt, readyAt, ok := calendar.Entries[0].OccupiedRange()
	if !ok || startAt.Format("2006-01-02 15:04") != "2026-12-31 08:00" || readyAt.Format("2006-01-02 15:04") != "2027-01-01 02:00" {
		t.Fatalf("cross-month occupancy range = %v to %v (valid %t), want 2026-12-31 08:00 to 2027-01-01 02:00", startAt, readyAt, ok)
	}
	ownerCalendar, err := NewCalendarService(conn).OwnerCalendar(1, 2027, time.January)
	if err != nil {
		t.Fatal(err)
	}
	if len(ownerCalendar.Entries) != 1 || ownerCalendar.Entries[0].GuestName != "Huésped fin de mes" || ownerCalendar.Entries[0].GuestPhone != "573001234567" || ownerCalendar.Entries[0].GuestEmail != "mes@example.com" {
		t.Fatalf("owner calendar should show reservation guest details, got %+v", ownerCalendar.Entries)
	}
}

func cabinTestDatabase(t *testing.T) (*sql.DB, db.ConnFn) {
	t.Helper()
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.SetMaxOpenConns(1)
	if err := db.RunMigrations(migrations.All, database); err != nil {
		database.Close()
		t.Fatal(err)
	}
	return database, func() (*sql.DB, error) { return database, nil }
}

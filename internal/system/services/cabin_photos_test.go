package services

import (
	"database/sql"
	"io/fs"
	"testing"
	"time"

	"travelando/internal/migrations"
	"travelando/internal/system/models"
)

func TestCabinPhotoMigrationPreservesCoverAndLoadsPhotosInOrder(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)

	legacyMigrations := migrationSQL(t, "20261105130000_add_cabin_photos.sql", "20261105140000_add_global_deposit.sql")
	for _, migration := range legacyMigrations {
		if _, err := database.Exec(migration); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := database.Exec(`
		INSERT INTO owners (name, whatsapp, pin_hash, cedula)
		VALUES ('Dueño', '573001234567', 'hash', '123456');
		INSERT INTO cabins (owner_id, name, cover_image, included_guests, day_price, night_price)
		VALUES (1, 'Cabaña', 'https://example.com/portada.jpg', 2, 100000, 200000);
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(migrationContents(t, "20261105130000_add_cabin_photos.sql")); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO cabin_photos (cabin_id, photo_url, position) VALUES (1, 'https://example.com/segunda.jpg', 1)`); err != nil {
		t.Fatal(err)
	}

	cabin, err := NewCabinService(func() (*sql.DB, error) { return database, nil }).Cabin(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(cabin.Photos) != 2 || cabin.Photos[0].URL != "https://example.com/portada.jpg" || cabin.Photos[1].URL != "https://example.com/segunda.jpg" {
		t.Fatalf("photos = %+v, want migrated cover then second photo", cabin.Photos)
	}

	adminPage, err := NewCabinService(func() (*sql.DB, error) { return database, nil }).CabinsPage(1, 10)
	if err != nil || len(adminPage.Cabins) != 1 || len(adminPage.Cabins[0].Photos) != 2 {
		t.Fatalf("admin cabin page photos = %+v, err %v", adminPage.Cabins, err)
	}
	ownerPage, err := NewOwnerService(func() (*sql.DB, error) { return database, nil }).OwnerCabinsPage(1, 1, 10)
	if err != nil || len(ownerPage.Cabins) != 1 || len(ownerPage.Cabins[0].Photos) != 2 {
		t.Fatalf("owner cabin page photos = %+v, err %v", ownerPage.Cabins, err)
	}
	from, _ := time.Parse(time.DateOnly, "2026-11-15")
	search, err := NewAvailabilityService(func() (*sql.DB, error) { return database, nil }).Search(SearchInput{
		StayType: models.StayDayUse, DateFrom: from, DateTo: from, Limit: 10,
	})
	if err != nil || len(search.Cabins) != 1 || len(search.Cabins[0].Photos) != 2 {
		t.Fatalf("guest search photos = %+v, err %v", search.Cabins, err)
	}

	rows, err := database.Query(`PRAGMA table_info(cabins)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, dataType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatal(err)
		}
		if name == "cover_image" {
			t.Fatal("legacy cover_image column remains")
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func migrationSQL(t *testing.T, excluded ...string) []string {
	t.Helper()
	exclude := make(map[string]bool, len(excluded))
	for _, name := range excluded {
		exclude[name] = true
	}
	files, err := fs.ReadDir(migrations.All, ".")
	if err != nil {
		t.Fatal(err)
	}
	var contentsList []string
	for _, file := range files {
		if file.Name() == "0_migrations.sql" || exclude[file.Name()] {
			continue
		}
		contents, err := fs.ReadFile(migrations.All, file.Name())
		if err != nil {
			t.Fatal(err)
		}
		contentsList = append(contentsList, string(contents))
	}
	return contentsList
}

func migrationContents(t *testing.T, name string) string {
	t.Helper()
	contents, err := fs.ReadFile(migrations.All, name)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

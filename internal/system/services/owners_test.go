package services

import (
	"errors"
	"testing"

	"travelando/internal/system/models"
)

func TestOwnerByIDReturnsOwnerProfileWithoutCredentials(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	result, err := database.Exec(`INSERT INTO owners (name, cedula, whatsapp, pin_hash) VALUES (?, ?, ?, ?)`, "María López", "12345678", "573001234567", "secret-hash")
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	owner, err := NewOwnerService(conn).OwnerByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if owner.ID != id || owner.Name != "María López" || owner.Cedula != "12345678" || owner.WhatsApp != "573001234567" {
		t.Fatalf("unexpected owner profile: %+v", owner)
	}
	if owner.HasPIN || owner.SessionVersion != 0 {
		t.Fatalf("profile exposed unnecessary auth state: %+v", owner)
	}
	if _, err := NewOwnerService(conn).OwnerByID(id + 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing owner error = %v, want ErrNotFound", err)
	}
}

func TestOwnerPINChangeInvalidatesExistingSession(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()
	svc := NewOwnerService(conn)
	owner := models.Owner{Name: "María López", Cedula: "12345678", WhatsApp: "573001234567"}
	if err := svc.CreateOwner(&owner, "123456"); err != nil {
		t.Fatal(err)
	}
	loggedIn, err := svc.AuthenticateOwner(owner.Cedula, "123456")
	if err != nil {
		t.Fatal(err)
	}

	updated := models.Owner{ID: owner.ID, Name: owner.Name, Cedula: owner.Cedula, WhatsApp: owner.WhatsApp}
	if err := svc.UpdateOwner(&updated, "654321"); err != nil {
		t.Fatal(err)
	}
	if valid, err := svc.OwnerSessionValid(owner.ID, loggedIn.SessionVersion); err != nil || valid {
		t.Fatalf("old owner session valid = %v, error = %v; want invalid", valid, err)
	}
	if _, err := svc.AuthenticateOwner(owner.Cedula, "123456"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("old PIN error = %v, want unauthorized", err)
	}
	if _, err := svc.AuthenticateOwner(owner.Cedula, "654321"); err != nil {
		t.Fatalf("new PIN did not authenticate: %v", err)
	}
}

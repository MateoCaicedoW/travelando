package services

import (
	"testing"

	"travelando/internal/system/models"
)

func TestValidateCabinCleaningHours(t *testing.T) {
	cabin := models.Cabin{Name: "Cabaña", Location: "Montaña", IncludedGuests: 1}
	if err := validateCabin(&cabin); err != nil {
		t.Fatalf("zero cleaning hours should be valid: %v", err)
	}
	cabin.CleaningHours = 3
	if err := validateCabin(&cabin); err != nil {
		t.Fatalf("three cleaning hours should be valid: %v", err)
	}
	cabin.CleaningHours = 25
	if err := validateCabin(&cabin); err == nil {
		t.Fatal("cleaning time over 24 hours should be rejected")
	}
}

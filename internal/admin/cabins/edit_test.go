package cabins

import (
	"net/http/httptest"
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestCabinEditResponseIsModalOnly(t *testing.T) {
	response := httptest.NewRecorder()
	data := cabinEditData{Cabin: models.Cabin{ID: 7, Name: "Cabaña"}}
	renderCabinEditDialog(response, data, nil, 200)
	if !strings.Contains(response.Body.String(), `id="modal"`) || strings.Contains(response.Body.String(), "<!DOCTYPE html>") {
		t.Fatalf("cabin edit rendered non-modal response: %s", response.Body.String())
	}
}

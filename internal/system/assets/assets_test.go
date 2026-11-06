package assets

import (
	"crypto/sha256"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestTermsPDFMatchesRecordedAcceptanceVersion(t *testing.T) {
	content, err := Files.ReadFile("terms-and-conditions.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256(content)); hash != models.CurrentTermsDocumentSHA256 {
		t.Fatalf("terms PDF SHA-256 = %s, want %s", hash, models.CurrentTermsDocumentSHA256)
	}

	response := httptest.NewRecorder()
	NewAssetHandler(Files, "/assets/").ServeHTTP(response, httptest.NewRequest("GET", "/assets/terms-and-conditions.pdf", nil))
	if response.Code != 200 || response.Header().Get("Content-Type") != "application/pdf" {
		t.Fatalf("terms PDF response = %d %q, want 200 application/pdf", response.Code, response.Header().Get("Content-Type"))
	}
}

func TestBrandIconsAreServedAsPNG(t *testing.T) {
	for _, name := range []string{
		"brand-orange-forest.png",
		"brand-orange-cream.png",
		"brand-cream-orange.png",
		"brand-forest-cream.png",
	} {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			NewAssetHandler(Files, "/assets/").ServeHTTP(response, httptest.NewRequest("GET", "/assets/"+name, nil))
			if response.Code != 200 || response.Header().Get("Content-Type") != "image/png" {
				t.Fatalf("icon response = %d %q, want 200 image/png", response.Code, response.Header().Get("Content-Type"))
			}
		})
	}
}

func TestComboboxDropdownEscapesScrollableModal(t *testing.T) {
	css, err := Files.ReadFile("css/tailwind.css")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		".modal-content:has(.travelando-combobox [data-popover][aria-hidden=\"false\"]) {\n  overflow-y: visible;\n}",
		"#modal:has(.travelando-combobox [data-popover][aria-hidden=\"false\"]) {\n  overflow-x: clip;\n}",
		".travelando-combobox [data-popover][aria-hidden=\"true\"] {\n  display: none;",
	} {
		if !strings.Contains(string(css), want) {
			t.Errorf("combobox overflow styles missing %q", want)
		}
	}
}

func TestActionDropdownStylesDoNotAffectOtherPopovers(t *testing.T) {
	css, err := Files.ReadFile("css/tailwind.css")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(css), `@import "basecoat-css/components/popover.css";`) {
		t.Fatal("Basecoat popover styles must stay scoped so they do not affect comboboxes")
	}
	for _, want := range []string{
		".dropdown-menu > [data-popover][aria-hidden=\"true\"] {\n  display: none;\n}",
		".dropdown-menu > [data-popover][aria-hidden=\"false\"] {\n  display: block;\n}",
	} {
		if !strings.Contains(string(css), want) {
			t.Errorf("scoped action dropdown styles missing %q", want)
		}
	}
}

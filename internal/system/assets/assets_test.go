package assets

import (
	"crypto/sha256"
	"fmt"
	"net/http/httptest"
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

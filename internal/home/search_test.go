package home

import (
	"net/http/httptest"
	"testing"
)

func TestInitialSearchPagePrefillsDateWithoutSearching(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	data := searchPageData(request, nil)

	if data.DateFrom == "" {
		t.Fatal("initial search date should be prefilled")
	}
	if data.DidSearch || data.Error != "" || len(data.Cabins) != 0 {
		t.Fatalf("initial page should not search or show results: %+v", data)
	}
}

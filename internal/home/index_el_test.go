package home

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestHomeCabinCardShowsPricesForAcceptedStayTypes(t *testing.T) {
	for _, test := range []struct {
		name             string
		stayType         models.StayType
		acceptsDayUse    bool
		acceptsOvernight bool
		dateFrom         string
		dateTo           string
		guests           int
		want             []string
		dontWant         []string
	}{
		{
			name: "both modalities while searching day use", stayType: models.StayDayUse,
			acceptsDayUse: true, acceptsOvernight: true, dateFrom: "2026-11-15", guests: 2,
			want:     []string{"$150.000", "Pasadía", "$250.000", "1 noche · incluye depósito", "grid grid-cols-2", `target="_blank"`, `rel="noopener noreferrer"`},
			dontWant: []string{"Vista al valle", "BBQ", "08:00 a 18:00"},
		},
		{
			name: "both modalities while searching overnight", stayType: models.StayOvernight,
			acceptsDayUse: true, acceptsOvernight: true, dateFrom: "2026-11-15", dateTo: "2026-11-17", guests: 2,
			want: []string{"$450.000", "Estadía completa · incluye depósito", "$150.000", "Pasadía", "grid grid-cols-2"},
		},
		{
			name: "only accepted modality", stayType: models.StayDayUse,
			acceptsDayUse: true, dateFrom: "2026-11-15", guests: 2,
			want: []string{"$150.000", "Pasadía", "grid grid-cols-1"}, dontWant: []string{"1 noche", "$250.000"},
		},
		{
			name: "overnight only", stayType: models.StayOvernight,
			acceptsOvernight: true, dateFrom: "2026-11-15", dateTo: "2026-11-17", guests: 2,
			want: []string{"$450.000", "Estadía completa · incluye depósito", "grid grid-cols-1"}, dontWant: []string{"$150.000"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := PageData{
				StayType: test.stayType, DateFrom: test.dateFrom, DateTo: test.dateTo, Guests: test.guests, DepositCOP: 50000,
				Cabins: []models.Cabin{{
					ID: 1, Name: "Cabaña", Location: "Montaña", Description: "Vista al valle", DayPrice: 100000, NightPrice: 200000,
					IncludedGuests: 2, AcceptsDayUse: test.acceptsDayUse, AcceptsOvernight: test.acceptsOvernight,
					DayStart: "08:00", DayEnd: "18:00", Amenities: []models.CabinAmenity{{Name: "BBQ", ValueType: models.AmenityBoolean, BooleanValue: true}},
				}},
				DidSearch: true, Page: 1,
			}
			var html strings.Builder
			if err := IndexEl(data).Render(&html); err != nil {
				t.Fatal(err)
			}
			rendered := html.String()
			start := strings.Index(rendered, "<article")
			if start < 0 {
				t.Fatalf("home result card not rendered: %s", rendered)
			}
			end := strings.Index(rendered[start:], "</article>")
			if end < 0 {
				t.Fatalf("home result card not rendered: %s", rendered)
			}
			card := rendered[start : start+end+len("</article>")]
			for _, want := range test.want {
				if !strings.Contains(card, want) {
					t.Errorf("home cabin card missing %q: %s", want, card)
				}
			}
			for _, dontWant := range test.dontWant {
				if strings.Contains(card, dontWant) {
					t.Errorf("home cabin card unexpectedly contains %q: %s", dontWant, card)
				}
			}
		})
	}
}

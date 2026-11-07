package home

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestHomeHeroUsesCatalogPhotoWithoutClaimingAvailabilityOrPrice(t *testing.T) {
	var html strings.Builder
	data := PageData{
		FeaturedCabin: &models.Cabin{
			Name: "Cabaña del río", Location: "San Francisco",
			Photos: []models.CabinPhoto{{URL: "/fotos/portada.jpg"}},
		},
		StayType: models.StayDayUse, Guests: 2,
	}
	if err := IndexEl(data).Render(&html); err != nil {
		t.Fatal(err)
	}
	rendered := html.String()
	for _, want := range []string{
		`src="/assets/hero-home.png"`,
		`type="radio"`, `name="stay_type"`,
		"Buscar cabañas disponibles",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("home hero missing %q: %s", want, rendered)
		}
	}
	for _, dontWant := range []string{"$100.000", "$200.000", "Disponible hoy"} {
		if strings.Contains(rendered, dontWant) {
			t.Errorf("unsearched home shows %q", dontWant)
		}
	}
}

func TestHomeHeroFallsBackWhenCatalogHasNoPhoto(t *testing.T) {
	var html strings.Builder
	if err := IndexEl(PageData{}).Render(&html); err != nil {
		t.Fatal(err)
	}
	rendered := html.String()
	if !strings.Contains(rendered, "Tu próximo plan está") || !strings.Contains(rendered, `src="/assets/hero-home.png"`) {
		t.Fatalf("home without cabin photos should keep search copy and omit broken image: %s", rendered)
	}
}

func TestHomeContainersUseCompactMobilePadding(t *testing.T) {
	var html strings.Builder
	if err := IndexEl(PageData{}).Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`class="mx-auto grid max-w-7xl items-center gap-5 px-2 py-2 sm:px-8 sm:py-10 lg:grid-cols-[1.05fr_0.95fr] lg:gap-8 lg:py-12"`,
		`class="mx-auto w-full max-w-7xl px-2 py-2 sm:px-8 sm:py-16"`,
	} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("home container missing compact mobile padding %q", want)
		}
	}
}

func TestHomeShowsThreeSampleCabinsBeforeSearchWithoutPricesOrAvailabilityClaims(t *testing.T) {
	data := PageData{SampleCabins: []models.Cabin{
		{ID: 1, Name: "Cabaña uno", Location: "Montaña", IncludedGuests: 2, DayPrice: 100000, AcceptsDayUse: true, Photos: []models.CabinPhoto{{URL: "/fotos/uno.jpg"}}},
		{ID: 2, Name: "Cabaña dos", Location: "Río", IncludedGuests: 4, NightPrice: 200000, AcceptsOvernight: true, Photos: []models.CabinPhoto{{URL: "/fotos/dos.jpg"}}},
		{ID: 3, Name: "Cabaña tres", Location: "Bosque", IncludedGuests: 3, DayPrice: 150000, NightPrice: 250000, AcceptsDayUse: true, AcceptsOvernight: true, Photos: []models.CabinPhoto{{URL: "/fotos/tres.jpg"}}},
	}}
	var html strings.Builder
	if err := IndexEl(data).Render(&html); err != nil {
		t.Fatal(err)
	}
	rendered := html.String()
	if got := strings.Count(rendered, "<article"); got != 3 {
		t.Fatalf("home shows %d cabin cards before search, want 3: %s", got, rendered)
	}
	for _, want := range []string{
		"Cabañas para empezar a explorar", "La disponibilidad se confirma al buscar tus fechas.",
		"Cabaña uno", "Cabaña dos", "Cabaña tres",
		`href="/cabins/1"`, `href="/cabins/2"`, `href="/cabins/3"`,
		"Pasadía", "Estadía", "Pasadía y estadía",
		`class="mt-12 grid gap-6 border-t border-line py-8 sm:mt-16 sm:grid-cols-3 sm:py-10"`,
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("home sample cabins missing %q", want)
		}
	}
	for _, dontWant := range []string{"$100.000", "$200.000", "$150.000", "Disponible hoy"} {
		if strings.Contains(rendered, dontWant) {
			t.Errorf("unsearched sample cabins show %q", dontWant)
		}
	}
}

func TestHomeHidesSampleCabinsAfterSearch(t *testing.T) {
	var html strings.Builder
	data := PageData{DidSearch: true, SampleCabins: []models.Cabin{{ID: 1, Name: "Cabaña aleatoria"}}}
	if err := IndexEl(data).Render(&html); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html.String(), "Cabañas para empezar a explorar") || strings.Contains(html.String(), "Cabaña aleatoria") {
		t.Fatalf("sample cabins should be hidden after search: %s", html.String())
	}
}

func TestHomeSearchScrollsToResultsAfterSubmit(t *testing.T) {
	var html strings.Builder
	if err := IndexEl(PageData{}).Render(&html); err != nil {
		t.Fatal(err)
	}
	rendered := html.String()
	for _, want := range []string{
		`action="/#search-results"`,
		`hx-swap="outerHTML show:#search-results:top"`,
		`<section id="search-results"`,
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("home search should scroll to results after submit; missing %q", want)
		}
	}
}

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
			want:     []string{"$150.000", "Pasadía", "$250.000", "1 noche · incluye depósito", "grid grid-cols-1 gap-0 sm:grid-cols-2 sm:gap-3", `target="_blank"`, `rel="noopener noreferrer"`},
			dontWant: []string{"Vista al valle", "BBQ", "08:00 a 18:00"},
		},
		{
			name: "both modalities while searching overnight", stayType: models.StayOvernight,
			acceptsDayUse: true, acceptsOvernight: true, dateFrom: "2026-11-15", dateTo: "2026-11-17", guests: 2,
			want: []string{"$450.000", "Estadía completa · incluye depósito", "$150.000", "Pasadía", "grid grid-cols-1 gap-0 sm:grid-cols-2 sm:gap-3"},
		},
		{
			name: "only accepted modality", stayType: models.StayDayUse,
			acceptsDayUse: true, dateFrom: "2026-11-15", guests: 2,
			want: []string{"$150.000", "Pasadía", "grid grid-cols-1 gap-2"}, dontWant: []string{"1 noche", "$250.000"},
		},
		{
			name: "overnight only", stayType: models.StayOvernight,
			acceptsOvernight: true, dateFrom: "2026-11-15", dateTo: "2026-11-17", guests: 2,
			want: []string{"$450.000", "Estadía completa · incluye depósito", "grid grid-cols-1 gap-2"}, dontWant: []string{"$150.000"},
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

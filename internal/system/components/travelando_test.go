package components

import (
	"bytes"
	"strings"
	"testing"

	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
)

func TestCabinCardUsesSharedDetailsAndPhotoCarousel(t *testing.T) {
	cabin := models.Cabin{
		Name: "Cabaña", Location: "Montaña", Description: "Vista al valle", IncludedGuests: 4,
		Photos: []models.CabinPhoto{{URL: "/foto-1", Position: 0}, {URL: "/foto-2", Position: 1}},
	}
	var output bytes.Buffer
	if err := CabinCard(cabin, CabinCardOptions{
		Details: Text("08:00 a 18:00"), Price: Text("$100.000 COP"), Action: Text("Ver cabaña"),
	}).Render(&output); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	for _, want := range []string{
		"/foto-1", "/foto-2", "Fotos de Cabaña", "Foto anterior", "Foto siguiente",
		`data-controller="photo-carousel"`, `data-action="click-&gt;photo-carousel#scroll"`,
		`data-action="click-&gt;photo-carousel#open"`, `data-photo-carousel-index="0"`,
		"Cabaña", "Montaña", "Vista al valle", "4 huéspedes incluidos", "08:00 a 18:00", "$100.000 COP", "Ver cabaña",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("cabin card missing %q: %s", want, html)
		}
	}
}

func TestBrandUsesTravelandoIcon(t *testing.T) {
	var output bytes.Buffer
	if err := Brand().Render(&output); err != nil {
		t.Fatal(err)
	}
	if html := output.String(); !strings.Contains(html, `src="/assets/brand-orange-cream.png"`) || !strings.Contains(html, `alt=""`) || !strings.Contains(html, "Travelando") {
		t.Fatalf("brand should use icon with accessible wordmark: %s", html)
	}
}

func TestCabinPhotoCarouselShowsPlaceholderWithoutPhotos(t *testing.T) {
	var output bytes.Buffer
	if err := CabinPhotoCarousel("Cabaña", nil).Render(&output); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	if !strings.Contains(html, "Fotos próximamente") || strings.Contains(html, "data-carousel-direction") {
		t.Fatalf("empty gallery should show placeholder without controls: %s", html)
	}
}

func TestCabinScheduleUsesSelectedStayType(t *testing.T) {
	cabin := models.Cabin{DayStart: "08:00", DayEnd: "18:00", CheckIn: "15:00", CheckOut: "12:00"}
	for _, test := range []struct {
		stayType models.StayType
		want     string
	}{
		{models.StayDayUse, "08:00 a 18:00"},
		{models.StayOvernight, "Entrada 15:00 · salida 12:00"},
	} {
		if got := CabinSchedule(test.stayType, cabin); got != test.want {
			t.Errorf("schedule for %s = %q, want %q", test.stayType, got, test.want)
		}
	}
}

func TestCabinPricesOnlyShowAcceptedStayTypes(t *testing.T) {
	for _, test := range []struct {
		name             string
		acceptsDayUse    bool
		acceptsOvernight bool
		want             []string
		dontWant         []string
	}{
		{name: "day use only", acceptsDayUse: true, want: []string{"Pasadía", "$100.000"}, dontWant: []string{"Por noche", "$200.000"}},
		{name: "overnight only", acceptsOvernight: true, want: []string{"Por noche", "$200.000"}, dontWant: []string{"Pasadía", "$100.000"}},
		{name: "both", acceptsDayUse: true, acceptsOvernight: true, want: []string{"Pasadía", "$100.000", "Por noche", "$200.000"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			cabin := models.Cabin{
				AcceptsDayUse: test.acceptsDayUse, AcceptsOvernight: test.acceptsOvernight,
				DayPrice: 100000, NightPrice: 200000,
			}
			var output bytes.Buffer
			if err := CabinPrices(cabin).Render(&output); err != nil {
				t.Fatal(err)
			}
			for _, want := range test.want {
				if !strings.Contains(output.String(), want) {
					t.Errorf("prices missing %q: %s", want, output.String())
				}
			}
			for _, dontWant := range test.dontWant {
				if strings.Contains(output.String(), dontWant) {
					t.Errorf("prices unexpectedly contain %q: %s", dontWant, output.String())
				}
			}
		})
	}
}

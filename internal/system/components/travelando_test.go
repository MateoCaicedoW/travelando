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

func TestSiteHeaderHidesExploreButtonOnlyOnHome(t *testing.T) {
	var output bytes.Buffer
	if err := SiteHeader("/").Render(&output); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "Explorar") {
		t.Fatal("home header should not show the explore button")
	}

	output.Reset()
	if err := SiteHeader("/cabins/1").Render(&output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Explorar") {
		t.Fatal("cabin detail header should keep the explore button")
	}
}

func TestSiteHeaderHidesExploreButtonOnReportPage(t *testing.T) {
	var output bytes.Buffer
	if err := SiteHeader("/report/token").Render(&output); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "Explorar") {
		t.Fatal("report header should not show the explore button")
	}
	if !strings.Contains(output.String(), "Travelando") {
		t.Fatal("report header should retain the shared brand")
	}
}

func TestMobileHeadersUseCompactGutters(t *testing.T) {
	for name, header := range map[string]Node{
		"public": SiteHeader("/"),
		"admin":  AdminMobileBar(),
		"owner":  OwnerMobileBar(),
	} {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			if err := header.Render(&output); err != nil {
				t.Fatal(err)
			}
			if html := output.String(); !strings.Contains(html, "px-2") || strings.Contains(html, "px-5") {
				t.Fatalf("mobile header should use compact gutters: %s", html)
			}
		})
	}
}

func TestOwnerNavigationShowsProfileAndLogoutActions(t *testing.T) {
	var output bytes.Buffer
	if err := OwnerNavigation("/owner/profile").Render(&output); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	for _, want := range []string{
		`href="/owner/profile"`, `aria-label="Perfil"`, `action="/owner/logout"`, "bg-forest text-white",
		`aria-label="Cerrar sesión"`, `data-portal-navigation-target="logoutButton"`, "hidden flex h-screen", "mt-auto flex items-center", "justify-between", `aria-current="page"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("owner sidebar missing %q: %s", want, html)
		}
	}
	if strings.Contains(html, "dropdown-menu") || strings.Contains(html, "owner-account-menu") {
		t.Fatal("owner sidebar should not use a profile dropdown")
	}
	profile := strings.Index(html, `href="/owner/profile"`)
	profileEnd := strings.Index(html[profile:], "</a>")
	if profile < 0 || profileEnd < 0 || !strings.Contains(html[profile:profile+profileEnd], "<svg") || !strings.Contains(html[profile:profile+profileEnd], "rounded-full") || strings.Contains(html[profile:profile+profileEnd], "<span") {
		t.Fatal("profile should be a round, icon-only link")
	}
	logout := strings.Index(html, `action="/owner/logout"`)
	logoutEnd := strings.Index(html[logout:], "</form>")
	if logout < 0 || logoutEnd < 0 || !strings.Contains(html[logout:logout+logoutEnd], "<svg") || strings.Contains(html[logout:logout+logoutEnd], "<span") {
		t.Fatal("logout form should show icon without label")
	}
}

func TestPortalCollapseButtonSharesBrandRow(t *testing.T) {
	for name, navigation := range map[string]Node{
		"admin": AdminHeader("/admin/"),
		"owner": OwnerNavigation("/owner/"),
	} {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			if err := navigation.Render(&output); err != nil {
				t.Fatal(err)
			}
			html := output.String()
			start := strings.Index(html, `data-portal-navigation-target="headerRow"`)
			if start < 0 {
				t.Fatalf("portal header row missing: %s", html)
			}
			row := html[start:]
			brand := strings.Index(row, `data-portal-navigation-target="brandLabel"`)
			if brand < 0 {
				t.Errorf("brand missing from portal header row: %s", row)
				return
			}
			brandEnd := strings.Index(row[brand:], `</a>`)
			if brandEnd < 0 || !strings.HasPrefix(row[brand+brandEnd:], `</a><div data-portal-navigation-target="collapseRow"`) {
				t.Errorf("collapse button should follow brand in same header row: %s", row)
			}
		})
	}
}

func TestSiteFooterIncludesSocialLinks(t *testing.T) {
	var output bytes.Buffer
	if err := SiteFooter().Render(&output); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	for _, want := range []string{
		`href="https://api.whatsapp.com/message/MQDALY2E3MHQK1?autoload=1&amp;app_absent=0"`,
		`href="https://www.instagram.com/travelando.co/"`,
		`href="https://www.tiktok.com/@travelandoconvictor"`,
		`target="_blank"`, `rel="noopener noreferrer"`, `aria-label="WhatsApp"`, `aria-label="Instagram"`, `aria-label="TikTok"`,
		`<svg`, `<path`, `<circle`, `<rect`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("site footer missing %q: %s", want, html)
		}
	}
}

func TestSiteFooterExplainsTravelandoAndLinksOwnerPortal(t *testing.T) {
	var output bytes.Buffer
	if err := SiteFooter().Render(&output); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	for _, want := range []string{
		"cabañas para pasadía o estadía", "su disponibilidad en tus fechas", "WhatsApp",
		`href="/#buscar"`, "Buscar cabañas", `href="/owner/login"`, "Portal de propietarios",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("footer missing %q: %s", want, html)
		}
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

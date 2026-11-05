package system

import (
	"bytes"
	"strings"
	"testing"

	. "maragu.dev/gomponents"
	"travelando/internal/system/components"
)

func renderNode(t *testing.T, node Node) string {
	t.Helper()
	var output bytes.Buffer
	if err := node.Render(&output); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func renderLayout(t *testing.T, portal string) string {
	t.Helper()
	return renderNode(t, (Layout{Title: "Test", Portal: portal, Yield: Text("Contenido")}).New())
}

func TestAdminLayoutUsesAdminNavigationWithoutExploreChrome(t *testing.T) {
	html := renderLayout(t, PortalAdmin)
	for _, want := range []string{"Inicio", "Reservas", "Calendario", "Cabañas", "Amenidades", "Dueños", `aria-label="Cerrar sesión"`, `action="/admin/logout"`, `aria-label="Abrir navegación"`, `aria-controls="admin-sidebar"`, `data-controller="admin-navigation"`, `id="admin-sidebar"`, "sticky top-0 z-30"} {
		if !strings.Contains(html, want) {
			t.Errorf("admin layout missing %q", want)
		}
	}
	for _, unwanted := range []string{"Administrar", "Encuentra tu próxima escapada.", `href="/#buscar"`} {
		if strings.Contains(html, unwanted) {
			t.Errorf("admin layout unexpectedly contains %q", unwanted)
		}
	}
}

func TestOwnerLayoutUsesOwnerNavigationWithoutExploreChrome(t *testing.T) {
	html := renderLayout(t, PortalOwner)
	for _, want := range []string{"Calendario", "Reservas", "Mis cabañas", `aria-label="Cerrar sesión"`, `action="/owner/logout"`} {
		if !strings.Contains(html, want) {
			t.Errorf("owner layout missing %q", want)
		}
	}
	for _, unwanted := range []string{"Administrar", "Dueños", "Encuentra tu próxima escapada.", `href="/#buscar"`} {
		if strings.Contains(html, unwanted) {
			t.Errorf("owner layout unexpectedly contains %q", unwanted)
		}
	}
	if strings.Contains(html, `data-controller="admin-navigation"`) || strings.Contains(html, `id="admin-sidebar"`) {
		t.Fatal("owner layout unexpectedly contains admin sidebar")
	}
}

func TestExploreHeaderDoesNotExposeAdministration(t *testing.T) {
	var output bytes.Buffer
	if err := (Layout{Title: "Test", Yield: Text("Contenido")}).New().Render(&output); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "Administrar") || strings.Contains(output.String(), `href="/admin/"`) {
		t.Fatal("explore layout exposes administration navigation")
	}
	if !strings.Contains(output.String(), "Explorar") {
		t.Fatal("explore layout is missing its explore link")
	}
}

func TestBareLayoutHasNoHeaderOrFooter(t *testing.T) {
	html := renderLayout(t, PortalBare)
	for _, unwanted := range []string{"Travelando", "Explorar", "Encuentra tu próxima escapada."} {
		if strings.Contains(html, unwanted) {
			t.Errorf("bare layout unexpectedly contains %q", unwanted)
		}
	}
}

func TestPortalHeaderMarksCurrentPage(t *testing.T) {
	html := renderNode(t, components.AdminHeader("/admin/calendar"))
	if !strings.Contains(html, `href="/admin/calendar" aria-current="page"`) {
		t.Fatal("portal navigation does not mark current page")
	}
	html = renderNode(t, components.AdminHeader("/admin/reservations"))
	if !strings.Contains(html, `href="/admin/reservations" aria-current="page"`) {
		t.Fatal("reservations navigation does not mark current page")
	}
	html = renderNode(t, components.OwnerHeader("/owner/reservations"))
	if !strings.Contains(html, `href="/owner/reservations" aria-current="page"`) {
		t.Fatal("owner reservations navigation does not mark current page")
	}
}

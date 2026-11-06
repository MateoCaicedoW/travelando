package owners

import (
	"strings"
	"testing"

	"travelando/internal/system/models"
)

func TestOwnersListHasMobileCardLabels(t *testing.T) {
	var html strings.Builder
	if err := ownersEl([]models.Owner{{ID: 7, Name: "Ana Pérez", Cedula: "123", WhatsApp: "573001234567", CabinCount: 2, CommissionPaid: 10000, CommissionDue: 5000, CommissionCredit: 2000, Active: true}}, 0, nil).Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"responsive-data-table", `data-label="Dueño"`, `data-label="Cédula"`, `data-label="Celular / WhatsApp"`, `data-label="Cabañas"`, `data-label="Comisión pagada"`, `data-label="Pendiente"`, `data-label="Saldo a favor"`, `data-label="Estado"`, `data-label="Acciones"`, `class="dropdown-menu relative inline-flex"`, `aria-label="Acciones"`, `aria-haspopup="menu"`, `role="menu"`, `role="menuitem"`, "Ana Pérez", "573001234567", "Editar", "Desactivar"} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("owners list missing %q", want)
		}
	}
}

func TestOwnerInlineEditRemainsAvailableInResponsiveTable(t *testing.T) {
	var html strings.Builder
	owner := models.Owner{ID: 7, Name: "Ana Pérez", Cedula: "123", WhatsApp: "573001234567", HasPIN: true, Active: true}
	if err := ownerEditRowEl(owner, nil).Render(&html); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`id="owner-row-7"`, `id="owner-7-name"`, `id="owner-7-cedula"`, `hx-target="#owners-content"`, "Guardar", "Cancelar"} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("owner inline edit missing %q", want)
		}
	}
}

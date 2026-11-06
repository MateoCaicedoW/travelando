package profile

import (
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/models"

	"go.leapkit.dev/core/server"
)

func renderOwnerProfile(w http.ResponseWriter, r *http.Request, owner models.Owner, state *ownerProfileState, status int) {
	w.WriteHeader(status)
	page := system.Layout{Title: "Perfil del propietario | Travelando", Portal: system.PortalOwner, Yield: ownerProfileEl(owner, state)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

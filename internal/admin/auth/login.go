package auth

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"os"

	"travelando/internal/system/helpers"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server/session"
)

func Login(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	fieldErrors := form.Validate(r, validate.Fields(validate.Field("password", validate.Required("Ingresa la contraseña."))))
	if fieldErrors.HasAny() {
		renderLoginPage(w, r, http.StatusBadRequest, fieldErrors)
		return
	}
	configured := os.Getenv("ADMIN_PASSWORD")
	provided := r.FormValue("password")
	if subtle.ConstantTimeCompare([]byte(configured), []byte(provided)) != 1 {
		fieldErrors["password"] = append(fieldErrors["password"], errors.New("Contraseña incorrecta o acceso sin configurar."))
		renderLoginPage(w, r, http.StatusUnauthorized, fieldErrors)
		return
	}
	sess := session.FromCtx(r.Context())
	if sess == nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	sess.Values[sessionKey] = true
	helpers.Redirect(w, r, "/admin/")
}

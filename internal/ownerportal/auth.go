package ownerportal

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"go.leapkit.dev/core/server/session"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"
)

const ownerSessionKey = "travelando_owner_id"

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ownerID, version := ownerSession(r)
		if ownerID < 1 {
			helpers.Redirect(w, r, "/owner/login")
			return
		}
		svc := services.ServiceSetFromCtx(r.Context()).Owners
		if svc == nil {
			http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
			return
		}
		valid, err := svc.OwnerSessionValid(ownerID, version)
		if err != nil {
			slog.Error("validating owner session", "error", err)
			http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
			return
		}
		if !valid {
			if sess := session.FromCtx(r.Context()); sess != nil {
				delete(sess.Values, ownerSessionKey)
			}
			helpers.Redirect(w, r, "/owner/login")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func ownerIDFromSession(r *http.Request) int64 {
	id, _ := ownerSession(r)
	return id
}

func ownerSession(r *http.Request) (int64, int64) {
	sess := session.FromCtx(r.Context())
	if sess == nil {
		return 0, 0
	}
	parts := strings.Split(sessionString(sess.Values[ownerSessionKey]), ":")
	if len(parts) != 2 {
		return 0, 0
	}
	id, _ := strconv.ParseInt(parts[0], 10, 64)
	version, _ := strconv.ParseInt(parts[1], 10, 64)
	return id, version
}

func sessionString(value any) string {
	valueString, _ := value.(string)
	return valueString
}

func ownerSessionValue(ownerID, version int64) string {
	return strconv.FormatInt(ownerID, 10) + ":" + strconv.FormatInt(version, 10)
}

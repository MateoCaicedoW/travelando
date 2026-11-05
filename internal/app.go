// Package internal integrates the app and loads general settings.
package internal

import (
	"cmp"
	"net/http"
	"os"

	adminamenities "travelando/internal/admin/amenities"
	adminauth "travelando/internal/admin/auth"
	admincabins "travelando/internal/admin/cabins"
	admincalendar "travelando/internal/admin/calendar"
	admindashboard "travelando/internal/admin/dashboard"
	adminowners "travelando/internal/admin/owners"
	"travelando/internal/home"
	"travelando/internal/ownerportal"
	"travelando/internal/requests"
	"travelando/internal/reservations"
	"travelando/internal/system/assets"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/db"
	"go.leapkit.dev/core/server"
	"go.leapkit.dev/core/server/session"
)

var (
	// DBFn is the database connection builder function
	// that will be used by the application based on the driver and
	// connection string.
	DBFn = db.ConnectionFn(
		cmp.Or(os.Getenv("DATABASE_URL"), "database.db"),

		db.WithDriver("sqlite3"),
		db.Params(
			"_busy_timeout", "5000",
			"_journal_mode", "WAL",
			"_sync", "1",
			"_cache_size", "8192",
			"_txlock", "deferred",
			"_foreign_keys", "1",

			// journal_size_limit
			// mmap_size
		),
	)

	// Server configuration variables loaded from
	host          = cmp.Or(os.Getenv("HOST"), "0.0.0.0")
	port          = cmp.Or(os.Getenv("PORT"), "3000")
	sessionName   = cmp.Or(os.Getenv("SESSION_NAME"), "leapkit_session")
	sessionSecret = cmp.Or(os.Getenv("SESSION_SECRET"), "change-me")
)

// New creates the http handler using the Leapkit server package
// and returns it with the address it is listening on.
func New() (http.Handler, string) {

	r := server.New(
		server.WithHost(host),
		server.WithPort(port),
		server.WithSession(sessionSecret, sessionName,
			session.WithHTTPOnly(true),
			session.WithSameSite(http.SameSiteLaxMode),
			session.WithSecureFlag(os.Getenv("GO_ENV") == "production"),
		),
	)
	r.Use(server.InCtxMiddleware(services.ServicesContextKey, services.NewServiceSet(DBFn)))

	r.HandleFunc("GET /{$}", home.Index)
	r.HandleFunc("GET /cabins/{id}", home.Cabin)
	r.HandleFunc("POST /requests/{id}", requests.Create)
	r.HandleFunc("GET /report/{token}", requests.Report)
	r.HandleFunc("POST /report/{token}", requests.ReportResult)
	r.HandleFunc("GET /admin/login", adminauth.LoginPage)
	r.HandleFunc("POST /admin/login", adminauth.Login)
	r.HandleFunc("GET /owner/login", ownerportal.LoginPage)
	r.HandleFunc("POST /owner/login", ownerportal.Login)

	r.Group("/admin", func(rg server.Router) {
		rg.Use(adminauth.RequireAuth)
		rg.HandleFunc("GET /{$}", admindashboard.Index)
		rg.HandleFunc("GET /inventory", admindashboard.Inventory)
		rg.HandleFunc("POST /commission", admindashboard.UpdateCommission)
		rg.HandleFunc("POST /inquiries/{id}/commission-paid", admindashboard.MarkCommissionPaid)
		rg.HandleFunc("GET /inquiries/{id}/edit", reservations.EditAdmin)
		rg.HandleFunc("POST /inquiries/{id}", reservations.UpdateAdmin)
		rg.HandleFunc("POST /logout", adminauth.Logout)
		rg.HandleFunc("GET /calendar", admincalendar.Index)
		rg.HandleFunc("GET /cabins", admincabins.Index)
		rg.HandleFunc("GET /cabins/new", admincabins.New)
		rg.HandleFunc("POST /cabins", admincabins.Create)
		rg.HandleFunc("GET /cabins/{id}/edit", admincabins.Edit)
		rg.HandleFunc("POST /cabins/{id}", admincabins.Update)
		rg.HandleFunc("POST /cabin-amenities/create", admincabins.CreateCabinAmenityTypeHandler)
		rg.HandleFunc("POST /cabins/{id}/blocks", admincabins.AddBlock)
		rg.HandleFunc("POST /cabins/{id}/blocks/{blockID}/delete", admincabins.DeleteBlock)
		rg.HandleFunc("GET /amenities", adminamenities.Index)
		rg.HandleFunc("POST /amenities/create", adminamenities.Create)
		rg.HandleFunc("POST /amenities/{id}", adminamenities.Update)
		rg.HandleFunc("POST /amenities/{id}/delete", adminamenities.Delete)
		rg.HandleFunc("GET /owners", adminowners.Index)
		rg.HandleFunc("GET /owners/new", adminowners.New)
		rg.HandleFunc("GET /owners/{id}/edit", adminowners.Edit)
		rg.HandleFunc("POST /owners", adminowners.Create)
		rg.HandleFunc("POST /owners/{id}/active", adminowners.SetActive)
		rg.HandleFunc("POST /owners/{id}", adminowners.Update)
	})
	r.Group("/owner", func(rg server.Router) {
		rg.Use(ownerportal.RequireAuth)
		rg.HandleFunc("GET /{$}", ownerportal.Index)
		rg.HandleFunc("GET /cabins", ownerportal.Cabins)
		rg.HandleFunc("GET /cabins/new", ownerportal.NewCabin)
		rg.HandleFunc("POST /cabins", ownerportal.CreateCabin)
		rg.HandleFunc("GET /cabins/{id}/edit", ownerportal.EditCabin)
		rg.HandleFunc("POST /cabins/{id}", ownerportal.UpdateCabin)
		rg.HandleFunc("POST /cabin-amenities/create", ownerportal.CreateCabinAmenityType)
		rg.HandleFunc("GET /amenities", ownerportal.AmenityTypes)
		rg.HandleFunc("POST /amenities/create", ownerportal.CreateAmenityType)
		rg.HandleFunc("POST /amenities/{id}", ownerportal.UpdateAmenityType)
		rg.HandleFunc("POST /amenities/{id}/delete", ownerportal.DeleteAmenityType)
		rg.HandleFunc("GET /blocks/new", ownerportal.NewBlock)
		rg.HandleFunc("POST /blocks", ownerportal.AddBlock)
		rg.HandleFunc("POST /blocks/{blockID}/delete", ownerportal.DeleteBlock)
		rg.HandleFunc("GET /reservations/{id}/edit", ownerportal.EditReservation)
		rg.HandleFunc("POST /reservations/{id}", ownerportal.UpdateReservation)
		rg.HandleFunc("POST /logout", ownerportal.Logout)
	})
	r.Handle("GET /assets/{path...}", assets.NewAssetHandler(assets.Files, "/assets/"))

	return r.Handler(), r.Addr()
}

# Agent notes

## Commands
- Go module targets Go 1.24; install Go and Node/npm before working.
- Local run: export required env vars, run `go run ./cmd/migrate`, then `go tool dev`; run `npm run dev` in a second terminal (Vite uses port 5173).
- `cmd/app` loads `.env`; `cmd/migrate` does not. Export `DATABASE_URL` before migrations if it comes from `.env`.
- Run `go test ./...`; focused examples: `go test ./internal -run TestNewRendersHomePage -count=1`, `go test ./internal/system -run TestLayoutUsesViteAssetsInProduction -count=1`.
- Run `npm run build` before `go build -o bin/app ./cmd/app`: Go embeds `internal/system/assets/**`; Vite writes production bundles to ignored `internal/system/assets/dist/`.
- Docker builds assets, app, and migrator, then runs `migrate && app`.

## Architecture
- `cmd/app/main.go` starts the server; `internal/app.go` wires routes, sessions, and services.
- Feature handlers live in `internal/home`, `internal/admin`, and `internal/requests`; larger features can use a nested feature directory (for example, `internal/admin/cabins`). Follow `.opencode/skills/travelando-feature/SKILL.md` for new features and feature refactors. Keep DB queries in `internal/system/services` and persisted types in `internal/system/models`.
- Gomponents markup lives in `*_el.go`; shared UI is in `internal/system/components`, shared utilities in `internal/system/helpers`.
- SQL migrations are embedded from `internal/migrations/*.sql`; add a new timestamped migration rather than editing an applied migration.

## Behavior that must stay aligned
- Availability ranges are half-open `[date_from, date_to)`, stored as ISO dates. Pasadía is stored as one day (`date_to` = next day); estadía ends on checkout. Keep search, inquiry creation, reservation resolution, and all cabin blocks consistent.
- Pending inquiries do not reserve dates. `reserved` inquiries and all `cabin_blocks` sources hide dates from search; owner-reported external bookings use `source = 'external'` and never create inquiries or commissions.
- Prices and totals are integer COP; commission rates are basis points. Reuse `services.Quote` and `services.Commission`, not float currency math.
- Owners log in with normalized cédula and six-digit PIN; PINs are PBKDF2 hashes, five failed logins lock access for 15 minutes, and cédula/PIN changes revoke sessions via `auth_version`. `cabins.owner_id` scopes owner access; cabin WhatsApp/contact comes from owner. Owners create cabins immediately and owner/admin can deactivate them. Owners may view and edit guest details only for reservations on their own cabins.
- Deactivating an owner revokes sessions and hides all their cabins from public availability without deleting cabins, reservations, or blocks. Reactivation restores only cabins that remain individually active.
- `commission_bps` snapshots the rate on reservation confirmation. `commission_paid_amount` tracks paid commission; reservation edits recalculate commission at the saved rate and append signed changes to `commission_adjustments`. Positive balance is outstanding; negative balance is owner credit. Owner summaries aggregate paid, outstanding, and credit. External blocks never create commission.
- The `report` token still grants access to one inquiry only and is visible in the guest's prefilled WhatsApp message; do not expand its access to other records.

## UI and configuration
- Read `PRODUCT.md` for product truth and `DESIGN.md` for visual tokens. Tailwind v4 scans Go files via `internal/system/assets/css/tailwind.css`; use its `@theme` color tokens rather than raw palette values.
- Homepage uses white/cream surfaces with orange and forest accents. Do not reintroduce `bg-ink` as the homepage background.
- Admin access uses `ADMIN_PASSWORD` (at least 12 chars) and production `SESSION_SECRET` (at least 32 chars); production `BASE_URL` must use HTTPS.
- Guest confirmation email uses implicit-TLS SMTP (`SMTP_PORT=465`). If SMTP is unavailable, the reservation is still saved.
- `.env` exists locally but is not ignored by `.gitignore`; never stage or print it.

---
name: Travelando Feature Pattern
description: Use when adding or refactoring a feature, endpoint, handler, service operation, or view in the Travelando Go app.
---

# Travelando feature pattern

Apply this workflow to every new feature and feature-level improvement.

## Before changing code

1. Read `AGENTS.md`; for UI changes also read `PRODUCT.md` and `DESIGN.md`.
2. Inspect the feature's routes, handlers, views, service calls, models, and existing callers before editing.
3. Keep scope to the requested feature and preserve current route and user-visible behavior unless the request changes it.

## Organize by feature and endpoint

- Keep feature code together. Admin cabins lives in `internal/admin/cabins/`; follow that nested feature pattern when a feature has its own area.
- Give each endpoint its own handler file, named for its action: `list.go`, `new.go`, `create.go`, `edit.go`, `update.go`, `delete.go`.
- Put a page's Gomponents view beside its handler with the same basename plus `_el.go`: `edit.go` → `edit_el.go`.
- Keep feature-specific shared markup in a clearly named feature file such as `cabin_form_el.go`. Do not duplicate a form just to create a view-file pair for POST endpoints; reuse the matching create/edit view.
- Register routes in `internal/app.go`; route handlers should be thin HTTP adapters that call feature functions. Keep business decisions, data aggregation, parsing, and persistence out of handlers.
- Name helper files for their responsibility (`form.go`, `edit_data.go`, `save.go`). Do not use a generic `service.go` as a catch-all for HTTP parsing, rendering, and unrelated helpers. Use `service.go` only for an actual service abstraction or implementation.

## Keep boundaries explicit

- Keep database queries and domain validation in `internal/system/services`; persisted types stay in `internal/system/models`.
- Feature-specific use cases or data aggregation can live in the feature package, in a descriptively named file.
- Reuse `services.RentalService` from context. Do not add a duplicate, feature-local interface unless the task explicitly requires a narrower boundary.
- Keep create and update operations explicit (`CreateCabin`, `UpdateCabin`); do not make one save operation infer intent from an ID.
- Validate parsed forms with `go.leapkit.dev/core/form.Validate` and `validate.Fields`; pass `validate.Errors` to the view and render each error below its input while preserving submitted values. The project pins core v0.1.16; API: https://github.com/leapkit/core/blob/v0.1.16/form/validate/validate.go.
- Put general redirect notices in the one-use session flash; `system.Layout` renders the global toast. Use `helpers.SetFlash`; do not put messages in query parameters or pass flash strings through individual views.
- Put generic, reusable Gomponents in `internal/system/components`. Keep feature-only markup in the feature. Put cross-feature non-UI utilities in `internal/system/helpers`.
- Check and propagate errors; never discard service, database, render, or form-parsing errors silently.

## Verify

- Run `gofmt` on changed Go files and `go build ./...`.
- Follow the user's instructions about tests; if the user says not to add or run tests, do not add or run them.
- Do not add dependencies or migrations unless the requested feature needs them.

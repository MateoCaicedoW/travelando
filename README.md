# Travelando

Catálogo de cabañas con búsqueda por disponibilidad, solicitudes conectadas a WhatsApp y seguimiento privado de reservas. Construido con Go, LeapKit, gomponents, HTMX, Tailwind CSS y SQLite; sigue la estructura por features y servicios de [Gasto](https://github.com/MateoCaicedoW/gasto).

## Desarrollo local

Requisitos: Go 1.24+, Node.js y npm.

```sh
go mod download
npm install
```

Configura el acceso privado y dirección pública de la aplicación:

```sh
export ADMIN_PASSWORD='una-frase-secreta-de-al-menos-12-caracteres'
export SESSION_SECRET="$(openssl rand -hex 32)"
export BASE_URL=http://localhost:3000
```

Inicializa la base y arranca la aplicación:

```sh
go run ./cmd/migrate
go tool dev
```

En otra terminal:

```sh
npm run dev
```

La aplicación corre en `http://localhost:3000` y Vite en `http://localhost:5173`.

## Correo de confirmación al huésped

El dueño reporta la reserva desde el enlace incluido en el WhatsApp prellenado. Travelando guarda el valor final, bloquea las fechas, calcula la comisión y envía confirmación al huésped. Configura SMTP para habilitar el correo:

El enlace de reporte es limitado a una solicitud, pero también queda visible al huésped dentro del mensaje prellenado; hoy no requiere clave del dueño.

```sh
export SMTP_SERVER=smtp.example.com
export SMTP_PORT=465
export SMTP_USERNAME=...
export SMTP_PASSWORD=...
export SMTP_FROM=reservas@example.com
```

Si SMTP no está configurado o falla, la reserva se conserva y el enlace muestra el aviso para enviar la confirmación por otro medio.

## Migraciones y producción

Las migraciones viven en `internal/migrations` y se ejecutan con `go run ./cmd/migrate`. Docker ejecuta migraciones al arrancar.

En producción configura `ADMIN_PASSWORD`, un `SESSION_SECRET` aleatorio, `BASE_URL` con HTTPS y las variables SMTP. Sin `SESSION_SECRET`, el panel permanece cerrado. Cada cabaña guarda su contacto y WhatsApp; la solicitud llega a ese número cuando el huésped pulsa **Enviar**. La comisión global se configura desde el panel y comienza en 0%.

## Verificación

```sh
gofmt -w ./cmd ./internal
go test ./...
npm run build
go build -o bin/app ./cmd/app
```

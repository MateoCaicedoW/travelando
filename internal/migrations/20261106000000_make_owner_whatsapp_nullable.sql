CREATE TABLE owners_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL CHECK(length(trim(name)) > 0),
    cedula TEXT NOT NULL DEFAULT '',
    whatsapp TEXT,
    pin_hash TEXT NOT NULL,
    failed_attempts INTEGER NOT NULL DEFAULT 0 CHECK(failed_attempts >= 0),
    locked_until INTEGER NOT NULL DEFAULT 0,
    auth_version INTEGER NOT NULL DEFAULT 1 CHECK(auth_version > 0),
    active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO owners_new (id, name, cedula, whatsapp, pin_hash, failed_attempts, locked_until, auth_version, active, created_at)
SELECT id, name, cedula, NULLIF(trim(whatsapp), ''), pin_hash, failed_attempts, locked_until, auth_version, active, created_at FROM owners;
DROP TABLE owners;
ALTER TABLE owners_new RENAME TO owners;
CREATE UNIQUE INDEX owners_whatsapp_unique ON owners(whatsapp) WHERE whatsapp IS NOT NULL;
CREATE UNIQUE INDEX owners_by_cedula ON owners(cedula) WHERE cedula <> '';
CREATE TABLE owners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL CHECK(length(trim(name)) > 0),
    whatsapp TEXT NOT NULL UNIQUE,
    pin_hash TEXT NOT NULL,
    failed_attempts INTEGER NOT NULL DEFAULT 0 CHECK(failed_attempts >= 0),
    locked_until INTEGER NOT NULL DEFAULT 0,
    auth_version INTEGER NOT NULL DEFAULT 1 CHECK(auth_version > 0),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE cabins ADD COLUMN owner_id INTEGER REFERENCES owners(id) ON DELETE SET NULL;
CREATE INDEX cabins_by_owner ON cabins(owner_id);

ALTER TABLE cabin_blocks ADD COLUMN source TEXT NOT NULL DEFAULT 'manual' CHECK(source IN ('manual', 'external'));

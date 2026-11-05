CREATE TABLE owners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL CHECK(length(trim(name)) > 0),
    whatsapp TEXT NOT NULL,
    commission_bps INTEGER NOT NULL CHECK(commission_bps BETWEEN 0 AND 10000),
    active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cabins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL REFERENCES owners(id),
    name TEXT NOT NULL CHECK(length(trim(name)) > 0),
    description TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    location_url TEXT NOT NULL DEFAULT '',
    cover_image TEXT NOT NULL DEFAULT '',
    max_guests INTEGER NOT NULL CHECK(max_guests > 0),
    day_price INTEGER NOT NULL CHECK(day_price BETWEEN 0 AND 1000000000000),
    night_price INTEGER NOT NULL CHECK(night_price BETWEEN 0 AND 1000000000000),
    day_start TEXT NOT NULL DEFAULT '09:00',
    day_end TEXT NOT NULL DEFAULT '18:00',
    check_in TEXT NOT NULL DEFAULT '15:00',
    check_out TEXT NOT NULL DEFAULT '12:00',
    active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inquiries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cabin_id INTEGER NOT NULL REFERENCES cabins(id),
    guest_name TEXT NOT NULL CHECK(length(trim(guest_name)) > 0),
    guest_phone TEXT NOT NULL,
    guest_email TEXT NOT NULL,
    stay_type TEXT NOT NULL CHECK(stay_type IN ('day_use', 'overnight')),
    date_from TEXT NOT NULL,
    date_to TEXT NOT NULL CHECK(date_to > date_from),
    guests INTEGER NOT NULL CHECK(guests > 0),
    estimated_total INTEGER NOT NULL CHECK(estimated_total >= 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'reserved', 'lost')),
    final_total INTEGER CHECK(final_total >= 0),
    commission_bps INTEGER CHECK(commission_bps BETWEEN 0 AND 10000),
    commission_due INTEGER CHECK(commission_due >= 0),
    report_token_hash TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TEXT
);

CREATE INDEX inquiries_by_cabin_dates ON inquiries(cabin_id, status, date_from, date_to);
CREATE INDEX inquiries_by_status ON inquiries(status, created_at DESC);

CREATE TABLE cabin_blocks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cabin_id INTEGER NOT NULL REFERENCES cabins(id) ON DELETE CASCADE,
    date_from TEXT NOT NULL,
    date_to TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK(date_to > date_from)
);

CREATE TRIGGER prevent_overlapping_reservations
BEFORE UPDATE OF status ON inquiries
WHEN NEW.status = 'reserved'
BEGIN
    SELECT CASE WHEN EXISTS (
        SELECT 1 FROM inquiries i
        WHERE i.cabin_id = NEW.cabin_id
          AND i.id <> NEW.id
          AND i.status = 'reserved'
          AND i.date_from < NEW.date_to
          AND i.date_to > NEW.date_from
    ) THEN RAISE(ABORT, 'dates unavailable') END;

    SELECT CASE WHEN EXISTS (
        SELECT 1 FROM cabin_blocks b
        WHERE b.cabin_id = NEW.cabin_id
          AND b.date_from < NEW.date_to
          AND b.date_to > NEW.date_from
    ) THEN RAISE(ABORT, 'dates unavailable') END;
END;

CREATE TRIGGER prevent_blocks_over_reservations
BEFORE INSERT ON cabin_blocks
BEGIN
    SELECT CASE WHEN EXISTS (
        SELECT 1 FROM inquiries i
        WHERE i.cabin_id = NEW.cabin_id
          AND i.status = 'reserved'
          AND i.date_from < NEW.date_to
          AND i.date_to > NEW.date_from
    ) THEN RAISE(ABORT, 'dates unavailable') END;
END;

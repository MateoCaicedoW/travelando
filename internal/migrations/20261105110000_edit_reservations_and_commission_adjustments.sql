ALTER TABLE inquiries ADD COLUMN commission_paid_amount INTEGER NOT NULL DEFAULT 0
    CHECK(commission_paid_amount BETWEEN 0 AND 1000000000000);

UPDATE inquiries
SET commission_paid_amount = commission_due
WHERE commission_paid_at IS NOT NULL AND commission_due IS NOT NULL;

CREATE TABLE commission_adjustments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    inquiry_id INTEGER NOT NULL REFERENCES inquiries(id),
    previous_amount INTEGER NOT NULL CHECK(previous_amount BETWEEN 0 AND 1000000000000),
    new_amount INTEGER NOT NULL CHECK(new_amount BETWEEN 0 AND 1000000000000),
    amount INTEGER NOT NULL CHECK(amount BETWEEN -1000000000000 AND 1000000000000),
    actor_role TEXT NOT NULL CHECK(actor_role IN ('admin', 'owner')),
    actor_owner_id INTEGER REFERENCES owners(id),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX commission_adjustments_by_inquiry ON commission_adjustments(inquiry_id, created_at);

DROP TRIGGER IF EXISTS prevent_overlapping_reservations;
CREATE TRIGGER prevent_overlapping_reservations
BEFORE UPDATE OF cabin_id, stay_type, date_from, date_to, start_time, end_time, cleaning_hours, status ON inquiries
WHEN NEW.status = 'reserved'
BEGIN
    SELECT CASE WHEN EXISTS (
        SELECT 1 FROM inquiries i
        WHERE i.cabin_id = NEW.cabin_id
          AND i.id <> NEW.id
          AND i.status = 'reserved'
          AND datetime(NEW.date_from || ' ' || NEW.start_time) < datetime(
              (CASE WHEN i.stay_type = 'day_use' THEN i.date_from ELSE i.date_to END) || ' ' || i.end_time,
              printf('+%d hours', i.cleaning_hours)
          )
          AND datetime(
              (CASE WHEN NEW.stay_type = 'day_use' THEN NEW.date_from ELSE NEW.date_to END) || ' ' || NEW.end_time,
              printf('+%d hours', NEW.cleaning_hours)
          ) > datetime(i.date_from || ' ' || i.start_time)
    ) THEN RAISE(ABORT, 'dates unavailable') END;

    SELECT CASE WHEN EXISTS (
        SELECT 1 FROM cabin_blocks b
        WHERE b.cabin_id = NEW.cabin_id
          AND b.date_from < NEW.date_to
          AND b.date_to > NEW.date_from
    ) THEN RAISE(ABORT, 'dates unavailable') END;
END;

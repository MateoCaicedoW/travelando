ALTER TABLE cabins ADD COLUMN cleaning_hours INTEGER NOT NULL DEFAULT 3 CHECK(cleaning_hours BETWEEN 0 AND 24);
ALTER TABLE inquiries ADD COLUMN start_time TEXT NOT NULL DEFAULT '09:00';
ALTER TABLE inquiries ADD COLUMN end_time TEXT NOT NULL DEFAULT '18:00';
ALTER TABLE inquiries ADD COLUMN cleaning_hours INTEGER NOT NULL DEFAULT 3 CHECK(cleaning_hours BETWEEN 0 AND 24);

UPDATE inquiries
SET start_time = (
        SELECT CASE WHEN inquiries.stay_type = 'day_use' THEN cabins.day_start ELSE cabins.check_in END
        FROM cabins WHERE cabins.id = inquiries.cabin_id
    ),
    end_time = (
        SELECT CASE WHEN inquiries.stay_type = 'day_use' THEN cabins.day_end ELSE cabins.check_out END
        FROM cabins WHERE cabins.id = inquiries.cabin_id
    ),
    cleaning_hours = (
        SELECT cabins.cleaning_hours FROM cabins WHERE cabins.id = inquiries.cabin_id
    );

DROP TRIGGER IF EXISTS prevent_overlapping_reservations;
CREATE TRIGGER prevent_overlapping_reservations
BEFORE UPDATE OF status ON inquiries
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

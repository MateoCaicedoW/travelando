ALTER TABLE cabins ADD COLUMN contact_name TEXT NOT NULL DEFAULT '';
ALTER TABLE cabins ADD COLUMN whatsapp TEXT NOT NULL DEFAULT '';
UPDATE cabins
SET contact_name = (SELECT name FROM owners WHERE owners.id = cabins.owner_id),
    whatsapp = (SELECT whatsapp FROM owners WHERE owners.id = cabins.owner_id),
    active = CASE WHEN active = 1 AND (SELECT active FROM owners WHERE owners.id = cabins.owner_id) = 1 THEN 1 ELSE 0 END;

ALTER TABLE cabins DROP COLUMN owner_id;
DROP TABLE owners;

CREATE TABLE commission_settings (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    commission_bps INTEGER NOT NULL DEFAULT 0 CHECK(commission_bps BETWEEN 0 AND 10000)
);

INSERT INTO commission_settings (id, commission_bps) VALUES (1, 0);

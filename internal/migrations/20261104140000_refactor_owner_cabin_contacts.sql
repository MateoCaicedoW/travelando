ALTER TABLE owners ADD COLUMN cedula TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX owners_by_cedula ON owners(cedula) WHERE cedula <> '';

INSERT OR IGNORE INTO owners (name, whatsapp, pin_hash)
SELECT COALESCE(MAX(NULLIF(trim(contact_name), '')), 'Dueño pendiente'), trim(whatsapp), ''
FROM cabins
WHERE owner_id IS NULL AND trim(whatsapp) <> ''
GROUP BY trim(whatsapp);

UPDATE cabins
SET owner_id = (SELECT id FROM owners WHERE owners.whatsapp = trim(cabins.whatsapp))
WHERE owner_id IS NULL AND trim(whatsapp) <> '';

ALTER TABLE cabins DROP COLUMN contact_name;
ALTER TABLE cabins DROP COLUMN whatsapp;
ALTER TABLE inquiries ADD COLUMN commission_paid_at TEXT;

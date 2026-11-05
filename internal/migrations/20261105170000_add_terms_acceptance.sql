ALTER TABLE inquiries ADD COLUMN terms_accepted_at TEXT;
ALTER TABLE inquiries ADD COLUMN terms_document_sha256 TEXT NOT NULL DEFAULT '';

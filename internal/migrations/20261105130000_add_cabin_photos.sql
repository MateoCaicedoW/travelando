CREATE TABLE cabin_photos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cabin_id INTEGER NOT NULL REFERENCES cabins(id) ON DELETE CASCADE,
    photo_url TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0 CHECK(position >= 0),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX cabin_photos_by_cabin ON cabin_photos(cabin_id, position, id);

INSERT INTO cabin_photos (cabin_id, photo_url, position)
SELECT id, cover_image, 0 FROM cabins WHERE trim(cover_image) <> '';

ALTER TABLE cabins DROP COLUMN cover_image;

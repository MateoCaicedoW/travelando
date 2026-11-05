ALTER TABLE cabins ADD COLUMN accepts_day_use INTEGER NOT NULL DEFAULT 1 CHECK(accepts_day_use IN (0, 1));
ALTER TABLE cabins ADD COLUMN accepts_overnight INTEGER NOT NULL DEFAULT 1 CHECK(accepts_overnight IN (0, 1));

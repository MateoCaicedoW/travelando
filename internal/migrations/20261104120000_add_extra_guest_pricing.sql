ALTER TABLE cabins RENAME COLUMN max_guests TO included_guests;
ALTER TABLE cabins ADD COLUMN extra_guest_price INTEGER NOT NULL DEFAULT 0
    CHECK(extra_guest_price BETWEEN 0 AND 1000000000000);

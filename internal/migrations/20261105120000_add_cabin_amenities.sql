CREATE TABLE cabin_amenity_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER REFERENCES owners(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK(length(trim(name)) BETWEEN 1 AND 80),
    value_type TEXT NOT NULL CHECK(value_type IN ('boolean', 'quantity')),
    CHECK(owner_id IS NULL OR owner_id > 0)
);

CREATE UNIQUE INDEX cabin_amenity_types_global_name ON cabin_amenity_types(lower(name)) WHERE owner_id IS NULL;
CREATE UNIQUE INDEX cabin_amenity_types_owner_name ON cabin_amenity_types(owner_id, lower(name)) WHERE owner_id IS NOT NULL;

CREATE TABLE cabin_amenities (
    cabin_id INTEGER NOT NULL REFERENCES cabins(id) ON DELETE CASCADE,
    amenity_type_id INTEGER NOT NULL REFERENCES cabin_amenity_types(id) ON DELETE CASCADE,
    boolean_value INTEGER CHECK(boolean_value IN (0, 1)),
    quantity_value INTEGER CHECK(quantity_value >= 0),
    PRIMARY KEY(cabin_id, amenity_type_id),
    CHECK(
        (boolean_value IS NOT NULL AND quantity_value IS NULL)
        OR (boolean_value IS NULL AND quantity_value IS NOT NULL)
    )
);

INSERT INTO cabin_amenity_types (name, value_type) VALUES
    ('Baños', 'quantity'),
    ('BBQ', 'boolean'),
    ('Mesa de billar', 'boolean');

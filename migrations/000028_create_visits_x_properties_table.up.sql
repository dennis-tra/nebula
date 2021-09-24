-- Begin the transaction
BEGIN;

-- The `visits_x_properties` table keeps track of
-- the most up to date association of a peer to its properties
CREATE TABLE visits_x_properties
(
    -- The visit ID of which we want to track the multi address
    visit_id    SERIAL,
    -- The ID of the property that with this visit is currently associated
    property_id SERIAL,

    -- The visit ID should always point to an existing visit in the DB
    CONSTRAINT fk_visits_x_properties_visit_id FOREIGN KEY (visit_id) REFERENCES visits (id) ON DELETE CASCADE,
    -- The property ID should always point to an existing property in the DB
    CONSTRAINT fk_visits_x_properties_property_id FOREIGN KEY (property_id) REFERENCES properties (id) ON DELETE CASCADE,

    PRIMARY KEY (visit_id, property_id)
);

CREATE INDEX idx_visits_x_properties_pkey_reversed ON visits_x_properties (property_id, visit_id);


-- End the transaction
COMMIT;

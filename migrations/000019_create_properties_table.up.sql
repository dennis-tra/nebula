-- The `properties` table holds information like agent versions, protocols or encountered errors.
CREATE TABLE properties
(
    -- An unique ID for this property
    id         SERIAL,
    -- The property name
    property   VARCHAR(500) NOT NULL,
    -- The property value
    value      VARCHAR(1000) NOT NULL,

    -- When was this property updated the last
    updated_at TIMESTAMPTZ  NOT NULL,
    -- When was this property created
    created_at TIMESTAMPTZ  NOT NULL,

    -- There should only be one unique property/value combination
    CONSTRAINT uq_properties_property_value UNIQUE (property, value),

    PRIMARY KEY (id)
);

-- Rows in the `ip_addresses` capture information for a particular IP address
-- that was derived from a multi address
CREATE TABLE ip_addresses
(
    -- A unique id that identifies this ip address
    id         SERIAL,
    -- When was this IP address updated
    updated_at TIMESTAMPTZ NOT NULL,
    -- When was this IP address created
    created_at TIMESTAMPTZ NOT NULL,
    -- the IP address
    address    INET        NOT NULL,
    -- The country that this address belongs to in the form of a two to three letter country code
    country    VARCHAR(2)  NOT NULL, -- make it not null so that the unique constraint applies IPs without country.

    -- Only one address/country combination should be allowed. If the IP address is assigned to a different
    -- country in the future this could be identified with this constraint.
    CONSTRAINT uq_address_country UNIQUE (address, country),

    PRIMARY KEY (id)
);

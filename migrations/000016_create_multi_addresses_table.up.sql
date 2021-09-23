-- Begin the transaction
BEGIN;

-- The `multi_addresses` table keeps track of all ever encountered multi addresses
-- some of these multi addresses can be associated with a country or cloud provider.
CREATE TABLE multi_addresses
(
    -- A unique id that identifies this multi_address
    id             SERIAL,
    -- The multi address in the form of `/ip4/123.456.789.123/tcp/4001`
    maddr          VARCHAR(200) NOT NULL,
    -- The derived IPv4 or IPv6 address that was used to determine the country/cloud provider
    addr           INET,
    -- The country that this multi address belongs to in the form of a two to three letter country code
    country        VARCHAR(3),
    -- The cloud provider that this multi address can be associated with
    cloud_provider VARCHAR(16),

    -- When was this multi address updated the last time
    updated_at     TIMESTAMPTZ  NOT NULL,
    -- When was this multi address created
    created_at     TIMESTAMPTZ  NOT NULL,

    -- There should only ever be distinct multi addresses here
    CONSTRAINT uq_multi_addresses_address UNIQUE (maddr),

    PRIMARY KEY (id)
);

CREATE INDEX idx_multi_addresses_maddr ON multi_addresses (maddr);

-- End the transaction
COMMIT;

BEGIN;

-- The `multi_addresses` table keeps track of all ever encountered multi addresses.
-- Some of these multi addresses can be associated with additional information.
CREATE TABLE multi_addresses
(
    -- An internal unique id that identifies this multi address.
    id             INT GENERATED ALWAYS AS IDENTITY,
    -- The autonomous system number that this multi address belongs to.
    asn            INT,
    -- If NULL this multi address could not be associated with a cloud provider.
    -- If not NULL the integer corresponds to the UdgerDB datacenter ID.
    is_cloud       INT,
    -- A boolean value that indicates whether this multi address is a relay address.
    is_relay       BOOLEAN,
    -- A boolean value that indicates whether this multi address is a publicly reachable one.
    is_public      BOOLEAN,
    -- The derived IPv4 or IPv6 address that was used to determine the country etc.
    addr           INET,
    -- Indicates if the multi_address has multiple IP addresses. Could happen for dnsaddr multi addresses.
    -- We moved the above IP address properties back to this table because of these numbers from a couple of months long
    -- crawler deployment:
    --     multi_address_count   ip_address_count
    --     896879                1
    --     133                   2
    --     2                     3
    --     1                     14
    -- This means the vast minority is only linked to multiple IP addresses.
    -- If this flag is true there are corresponding IP addresses.
    has_many_addrs BOOLEAN,
    -- The country that this multi address belongs to in the form of a two letter country code.
    country        CHAR(2),
    -- The continent that this multi address belongs to in the form of a two letter code.
    continent      CHAR(2),
    -- The multi address in the form of `/ip4/123.456.789.123/tcp/4001`.
    maddr          TEXT        NOT NULL,

    -- When was this multi address updated the last time
    updated_at     TIMESTAMPTZ NOT NULL,
    -- When was this multi address created
    created_at     TIMESTAMPTZ NOT NULL,

    -- There should only ever be distinct multi addresses here
    CONSTRAINT uq_multi_addresses_address UNIQUE (maddr),

    PRIMARY KEY (id)
);


COMMENT ON TABLE multi_addresses IS ''
    'The `multi_addresses` table keeps track of all ever encountered multi addresses.'
    'Some of these multi addresses can be associated with additional information.';
COMMENT ON COLUMN multi_addresses.id IS 'An internal unique id that identifies this multi address.';
COMMENT ON COLUMN multi_addresses.asn IS 'The autonomous system number that this multi address belongs to.';
COMMENT ON COLUMN multi_addresses.is_cloud IS 'If NULL this multi address could not be associated with a cloud provider. If not NULL the integer corresponds to the UdgerDB datacenter ID.';
COMMENT ON COLUMN multi_addresses.is_relay IS 'A boolean value that indicates whether this multi address is a relay address.';
COMMENT ON COLUMN multi_addresses.is_public IS 'A boolean value that indicates whether this multi address is a publicly reachable one.';
COMMENT ON COLUMN multi_addresses.addr IS 'The derived IPv4 or IPv6 address of this multi address.';
COMMENT ON COLUMN multi_addresses.has_many_addrs IS 'Indicates if the multi_address has multiple IP addresses. Could happen for dnsaddr multi addresses.';
COMMENT ON COLUMN multi_addresses.country IS 'The country that this multi address belongs to in the form of a two letter country code.';
COMMENT ON COLUMN multi_addresses.continent IS 'The continent that this multi address belongs to in the form of a two letter code.';
COMMENT ON COLUMN multi_addresses.maddr IS 'The multi address in the form of `/ip4/123.456.789.123/tcp/4001`.';
COMMENT ON COLUMN multi_addresses.updated_at IS 'Timestamp of when this multi address was updated.';
COMMENT ON COLUMN multi_addresses.created_at IS 'Timestamp of when this multi address was created.';


CREATE INDEX idx_multi_addresses_maddr ON multi_addresses (maddr);

COMMIT;

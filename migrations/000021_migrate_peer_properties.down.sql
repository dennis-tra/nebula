-- Begin the transaction
BEGIN;

-- Remove obsolete/redundant columns
ALTER TABLE crawl_properties
    ADD COLUMN property VARCHAR(255);
ALTER TABLE crawl_properties
    ADD COLUMN value VARCHAR(255);

ALTER TABLE crawl_properties
    DROP CONSTRAINT fk_crawl_properties_property;

UPDATE crawl_properties AS cp
SET property=subquery.property,
    value=subquery.value
FROM (SELECT id, property, value FROM properties) AS subquery
WHERE subquery.id = cp.property_id;

-- Now that we have filled the column we set it to be not null.
ALTER TABLE crawl_properties
    ALTER COLUMN property SET NOT NULL;
ALTER TABLE crawl_properties
    ALTER COLUMN value SET NOT NULL;

ALTER TABLE crawl_properties
    DROP COLUMN property_id;

DELETE
FROM properties;

-- End the transaction
COMMIT;

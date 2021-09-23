-- Begin the transaction
BEGIN;

-- Migrate existing peer properties from all crawls to crawl_properties table
INSERT INTO properties (property, value, updated_at, created_at)
SELECT property, value, NOW(), NOW()
FROM crawl_properties
GROUP BY property, value;

-- Add reference column
ALTER TABLE crawl_properties
    ADD COLUMN property_id SERIAL;

-- Fill reference column on crawl_crawl_properties table by letting
-- it point to the correct peer property in the respective table.
UPDATE crawl_properties AS pp
SET property_id=subquery.id
FROM (SELECT id, property, value FROM properties) AS subquery
WHERE subquery.property = pp.property
  AND subquery.value = pp.value;

-- Now that we have filled the column we set it to be not null.
ALTER TABLE crawl_properties
    ALTER COLUMN property_id SET NOT NULL;

-- Remove obsolete/redundant columns
ALTER TABLE crawl_properties DROP COLUMN property;
ALTER TABLE crawl_properties DROP COLUMN value;

-- Make sure the property_id always has a corresponding counterpart
ALTER TABLE crawl_properties
    ADD CONSTRAINT fk_crawl_properties_property
        FOREIGN KEY (property_id)
            REFERENCES properties (id)
            ON DELETE CASCADE;

-- End the transaction
COMMIT;

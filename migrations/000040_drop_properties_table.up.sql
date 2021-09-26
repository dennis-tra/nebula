-- Begin the transaction
BEGIN;

DROP TABLE visits_x_properties;
DROP TABLE peers_x_properties;
DROP TABLE properties;

-- End the transaction
COMMIT;

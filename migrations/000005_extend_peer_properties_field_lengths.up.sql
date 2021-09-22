-- Increase column sizes as agent version strings can become quite long.
ALTER TABLE peer_properties
    ALTER COLUMN value TYPE VARCHAR(500);
ALTER TABLE peer_properties
    ALTER COLUMN property TYPE VARCHAR(500);

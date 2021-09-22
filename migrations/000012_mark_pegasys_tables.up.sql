-- Rename tables from pegasys team as I want to work with a different schema
ALTER TABLE connections
    RENAME TO pegasys_connections;
ALTER TABLE neighbours
    RENAME TO pegasys_neighbours;

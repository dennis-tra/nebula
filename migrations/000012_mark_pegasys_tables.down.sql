-- Rename tables from pegasys team as I want to work with a different schema
ALTER TABLE pegasys_connections
    RENAME TO connections;
ALTER TABLE pegasys_neighbours
    RENAME TO neighbours;

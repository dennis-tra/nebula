-- After a crawl there may be the same peer ID with different multi addresses. This means
-- that the session stopped at one point and resumed e.g., on a different port. Right now
-- the multi addresses are just overwritten when we find a peer in the DHT and don't factor
-- that in for the session handling. With the additional column we can atomically upsert
-- the peer ID and return the old multi addresses to determine at the application level
-- if this was actually a new session or still the old one.
ALTER TABLE peers ADD COLUMN old_multi_addresses VARCHAR(255) ARRAY;

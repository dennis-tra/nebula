-- Fixes the error:
--   unable to insert into crawls: pq: duplicate key value violates unique constraint "crawls_pkey1"
SELECT setval(pg_get_serial_sequence('crawls', 'id'), (SELECT MAX(id) FROM crawls)+1)

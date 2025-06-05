-- DO NOT EDIT: This file was generated with: just generate-local-clickhouse-migrations

ALTER TABLE visits
    ADD COLUMN listen_maddrs Array(String) DEFAULT [] AFTER extra_maddrs;
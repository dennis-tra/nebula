ALTER TABLE visits
    ADD COLUMN listen_maddrs Array(String) DEFAULT [] AFTER extra_maddrs;
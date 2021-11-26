#!/bin/bash

set -e
set -u

echo "Creating multiple databases"

{% for network in networks %}

echo "Configuring database for the {{ network.name }} network"
echo "Creating database users"

psql -U "{{ db_user }}" <<-EOSQL
    CREATE USER nebula_{{ network.name }} WITH LOGIN PASSWORD '{{ network.db_password }}';
    CREATE USER nebula_{{ network.name }}_read WITH LOGIN PASSWORD '{{ network.db_password_read }}';
EOSQL

echo "Creating database for the {{ network.name }} network..."
psql --echo-all -U "{{ db_user }}" <<-EOSQL
    CREATE DATABASE nebula_{{ network.name }};
EOSQL

echo "Granting user permissions for the {{ network.name }} network..."
psql --echo-all -v ON_ERROR_STOP=1 -U "{{ db_user }}" -d nebula_{{ network.name }} <<-EOSQL
  GRANT ALL PRIVILEGES ON DATABASE nebula_{{ network.name }} TO nebula_{{ network.name }};
  GRANT CONNECT ON DATABASE nebula_{{ network.name }} TO nebula_{{ network.name }}_read;
  GRANT USAGE ON SCHEMA public TO nebula_{{ network.name }}_read;
  GRANT SELECT ON ALL TABLES IN SCHEMA public TO nebula_{{ network.name }}_read;
  ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO nebula_{{ network.name }}_read;
EOSQL

{% endfor %}

echo "Done!"

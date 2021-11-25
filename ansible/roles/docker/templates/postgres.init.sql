{% for network in networks %}
-- Create write user for the {{ network.name }} network
CREATE USER nebula_{{ network.name }} WITH LOGIN PASSWORD '{{ network.db_password }}';
CREATE DATABASE nebula_{{ network.name }};
GRANT ALL PRIVILEGES ON DATABASE nebula_{{ network.name }} TO nebula_{{ network.name }};

-- Create read-only user for the {{ network.name }} network
CREATE USER nebula_{{ network.name }}_read WITH LOGIN PASSWORD '{{ network.db_password_read }}';
\c nebula_{{ network.name }};
GRANT CONNECT ON DATABASE nebula_{{ network.name }} TO nebula_{{ network.name }}_read;
GRANT USAGE ON SCHEMA public TO nebula_{{ network.name }}_read;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO nebula_{{ network.name }}_read;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO nebula_{{ network.name }}_read;

{% endfor %}

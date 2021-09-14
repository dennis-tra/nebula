import psycopg2
import toml
import time_check

config = toml.load("./db.toml")['psql']
conn = psycopg2.connect(
    host=config['host'],
    port=config['port'],
    database=config['database'],
    user=config['user'],
    password=config['password'],
)

# Get the start, end time
start, end = time_check.get_time_range(conn)
print(start, end)

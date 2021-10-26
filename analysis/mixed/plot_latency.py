import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_reliability, node_geolocation, node_latency

config = toml.load("./db.toml")['psql']
conn = psycopg2.connect(
    host=config['host'],
    port=config['port'],
    database=config['database'],
    user=config['user'],
    password=config['password'],
)

start, end = node_time.get_time_range(conn)
all = node_classification.get_all_nodes(conn, start, end)
locations = node_geolocation.get_geolocation(conn, all)

latency_ms_list = node_latency.get_latency_list(conn)

res = dict()
for id, latency in latency_ms_list.items():
    if latency > 0:
        loc = locations.get(id, "unknown")
        if loc not in res:
            res[loc] = []
        res[loc].append(latency)

ress = dict()
for loc, latencies in res.items():
    ress[loc] = (sum(latencies) / len(latencies), len(latencies))
print(ress)

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

# print(locations)
node_latency.update_all_peers_latency(conn, locations)
location_list, latency_ms_list = node_latency.get_latency_list(conn)

# Plot
fig = plt.figure()
ax = fig.add_axes([0,0,1,1])

ax.bar(location_list, latency_ms_list)

plt.show()
import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification

config = toml.load("./db.toml")['psql']
conn = psycopg2.connect(
    host=config['host'],
    port=config['port'],
    database=config['database'],
    user=config['user'],
    password=config['password'],
)

start, end = node_time.get_time_range(conn)
dangle = node_classification.get_dangling_nodes(conn, start, end)
on = node_classification.get_on_nodes(conn, start, end)
off = node_classification.get_off_nodes(conn, start, end)

# Plot
plt.rc('font', size=8)
plt.pie([len(off), len(on), len(dangle)],
        labels=["off nodes %d" % len(off), "on nodes %d" % len(on), "dangling nodes %d" % len(dangle)],
        autopct="%.1f%%")
plt.title("Node classification from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))
plt.show()

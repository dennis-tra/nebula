import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_reliability

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
reliabilities = node_reliability.get_reliability(conn, start, end, dangle)

# Plot
plt.rc('font', size=8)
plt.hist(reliabilities, density=False, bins=50)
plt.title("Dangling nodes reliabilities from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))
plt.xlabel("Reliability in %")
plt.ylabel("Node counts")
plt.savefig("./figs/dangling_nodes_reliability.png")

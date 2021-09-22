import psycopg2
import toml
import matplotlib.pyplot as plt
import numpy as np
from lib import node_time, node_uptime

config = toml.load("./db.toml")['psql']
conn = psycopg2.connect(
    host=config['host'],
    port=config['port'],
    database=config['database'],
    user=config['user'],
    password=config['password'],
)

start, end = node_time.get_time_range(conn)
uptimes = [i for i in node_uptime.get_node_uptime(conn, start, end) if i]

# Plotting cdf of uptimes, code adapted from ../churn/cdf.py
hist_values, bin_edges = np.histogram(
    uptimes, bins=len(uptimes), density=True
)

# Since we provided an integer to the bins parameter above. The edges are equal width.
# This means the width between the first two elements is equal for all edges.
edge_width = bin_edges[1] - bin_edges[0]

# Integerate over histogram
cumsum = np.cumsum(hist_values) * edge_width

# build plot
plt.plot(bin_edges[1:], cumsum, label="All sessions")

# Presentation logic
plt.rc('font', size=8)
plt.gca().xaxis.set_major_formatter(lambda x, pos=None: x / 3600)
plt.xlabel("Hours")
plt.ylabel("Percentage of online peers")
plt.tight_layout()
plt.xticks(np.arange(0, max(bin_edges[1:]), 3 * 60 * 60))
plt.xlim(-60 * 60, 24 * 60 * 60)
plt.grid(True)
plt.legend()
plt.title("Session cdf from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))

# Finalize
plt.show()

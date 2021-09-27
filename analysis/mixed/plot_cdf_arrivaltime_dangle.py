import psycopg2
import toml
import matplotlib.pyplot as plt
import numpy as np
from lib import node_time, node_inter_arrival, node_classification

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
arrival = [i for i in node_inter_arrival.get_arrival_time(conn, start, end, dangle) if i]

# Plotting cdf of uptimes, code adapted from ../churn/cdf.py
hist_values, bin_edges = np.histogram(
    arrival, bins=len(arrival), density=True
)
edge_width = bin_edges[1] - bin_edges[0]
cumsum = np.cumsum(hist_values) * edge_width

plt.rc('font', size=8)
plt.plot(bin_edges[1:], cumsum, label="All inter arrival time")
plt.gca().xaxis.set_major_formatter(lambda x, pos=None: round(x / 3600, 2))
plt.xlabel("Hours")
plt.ylabel("CDF of inter arrival time")
plt.tight_layout()
plt.xticks(np.arange(0, max(bin_edges[1:]), 3 * 60 * 60))
plt.grid(True)
plt.legend()
plt.title("Dangling nodes inter-arrival time cdf from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))

# Finalize
plt.show()

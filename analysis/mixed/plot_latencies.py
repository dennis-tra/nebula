import psycopg2
import toml
import numpy as np
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_latency

# Note, sudo is required to run this script.
config = toml.load("./db.toml")['psql']
conn = psycopg2.connect(
    host=config['host'],
    port=config['port'],
    database=config['database'],
    user=config['user'],
    password=config['password'],
)

latencies = node_latency.get_latencies(conn)

lats = []
for latency in latencies:
    lats += [latency[0]]

print(len(latencies))

plt.hist(np.array(lats)*1000, bins=np.arange(0, 1000, 20))
# Plot
plt.xlabel("Ping latency in ms")
plt.ylabel("Fraction of peers")
plt.rc('font', size=8)
plt.show()

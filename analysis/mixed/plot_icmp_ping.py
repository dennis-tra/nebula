import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_ping
import os
import psutil

# Note, sudo is required to run this script.
config = toml.load("./db.toml")['psql']
conn = psycopg2.connect(
    host=config['host'],
    port=config['port'],
    database=config['database'],
    user=config['user'],
    password=config['password'],
)

start, end = node_time.get_time_range(conn)
on = node_classification.get_on_nodes(conn, start, end)
pings = node_ping.check_node_ping(conn, on)

counts = {"succeed": 0, "failed": 0}
sum = 0
for _, ping in pings.items():
    if ping:
        counts['succeed'] += 1
    else:
        counts['failed'] += 1

# Plot
plt.rc('font', size=8)
plt.pie(counts.values(), labels=counts.keys(), autopct="%.1f%%")
plt.title("On nodes ping successful from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))
plt.savefig("./figs/icmp_ping_nodes.png")
print("memory used:", psutil.Process(os.getpid()).memory_info().rss / 1024 ** 2, "MB")
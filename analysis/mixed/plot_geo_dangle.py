import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_geolocation
import os
import psutil


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
locations = node_geolocation.get_geolocation(conn, dangle)

counts = dict()
sum = 0
for _, location in locations.items():
    if location in counts:
        counts[location] += 1
    else:
        counts[location] = 1
    sum += 1
countsTrim = {"others": 0}
for key, val in counts.items():
    if val / sum < 0.01:
        countsTrim["others"] += val
    else:
        countsTrim[key] = val
{k: v for k, v in sorted(countsTrim.items(), key=lambda item: item[1])}

# Plot
plt.rc('font', size=8)
plt.pie(countsTrim.values(), labels=countsTrim.keys(), autopct="%.1f%%")
plt.title("Dangling nodes geolocation info from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))
plt.savefig("./figs/geolocation_for_dangling_nodes.png")
print("memory used:", psutil.Process(os.getpid()).memory_info().rss / 1024 ** 2, "MB")
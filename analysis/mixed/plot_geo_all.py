import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_geolocation
import numpy as np


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
colormap = plt.cm.nipy_spectral
colors = [colormap(i) for i in np.linspace(0, 1, len(countsTrim))]
patches, texts, _ = plt.pie(countsTrim.values(), labels=countsTrim.keys(), autopct="%.1f%%", colors=colors)
labels = []
for key in countsTrim.keys():
    labels.append("{}-{}".format(key, countsTrim[key]))
plt.legend(patches, labels, loc="best")
plt.title("All nodes geolocation info from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))
plt.show()

import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_cloud


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
clouds = node_cloud.get_cloud(conn, on)

counts = dict()
sum = 0
for _, cloud in clouds.items():
    if cloud in counts:
        counts[cloud] += 1
    else:
        counts[cloud] = 1

# Plot
plt.rc('font', size=8)
plt.pie(counts.values(), labels=counts.keys(), autopct="%.1f%%")
plt.title("On nodes cloud info from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))
plt.show()

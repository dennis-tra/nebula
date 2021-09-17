import pytz
import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_geolocation, node_correlation

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

hk = []
locations = node_geolocation.get_geolocation(conn, dangle)
for id, loc in locations.items():
    if loc == "HK":
        hk.append(id)

# Calculate correlation
corrOn, corrOff = node_correlation.get_up_time_correlation(conn, start, end, 600, 28800, 36000, hk, pytz.timezone("Asia/Hong_Kong"))

# Plot
plt.rc('font', size=8)
fig, axs = plt.subplots(1, 2)
fig.suptitle("Dangling nodes in HK uptime correlation distribution")
axs[0].hist(list(corrOn.values()), bins=20, density=False)
axs[0].set_title("Correlation with day time")
axs[0].set_xlabel("Correlation")
axs[0].set_ylabel("Node counts")
axs[1].hist(list(corrOff.values()), bins=20, density=False)
axs[1].set_title("Correlation with night time")
axs[1].set_xlabel("Correlation")
axs[1].set_ylabel("Node counts")
plt.show()
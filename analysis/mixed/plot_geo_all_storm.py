import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_agent, node_geolocation


# Helper function to trim agent version
def trim_agent(agent):
    if agent.startswith("/"):
        version = agent[1:]
    if agent.startswith("go-ipfs"):
        return "go-ipfs"
    elif agent.startswith("hydra-booster"):
        return "hydra-booster"
    elif agent.startswith("storm"):
        return "storm"
    elif agent.startswith("ioi"):
        return "ioi"
    else:
        return "others"


config = toml.load("./db.toml")['psql']
conn = psycopg2.connect(
    host=config['host'],
    port=config['port'],
    database=config['database'],
    user=config['user'],
    password=config['password'],
)

start, end = node_time.get_time_range(conn)
# Get storm node ids
all = node_classification.get_all_nodes(conn, start, end)
agents = node_agent.get_agent_version(conn, all)
storm = set()
for id, agent in agents.items():
    agent = trim_agent(agent)
    if agent == "storm":
        storm.add(id)

locations = node_geolocation.get_geolocation(conn, storm)

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
patches, texts, _ = plt.pie(countsTrim.values(), labels=countsTrim.keys(), autopct="%.1f%%")
labels = []
for key in countsTrim.keys():
    labels.append("{}-{}".format(key, countsTrim[key]))
plt.legend(patches, labels, loc="best")
plt.title("Storm nodes geolocation info from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))
plt.savefig("./figs/geolocation_for_all_storm.png")

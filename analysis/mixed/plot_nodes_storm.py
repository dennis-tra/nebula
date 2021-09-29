import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_agent


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

dangle = node_classification.get_dangling_nodes(conn, start, end)
dangleStorm = 0
for id in dangle:
    if id in storm:
        dangleStorm += 1
on = node_classification.get_on_nodes(conn, start, end)
onStorm = 0
for id in on:
    if id in storm:
        onStorm += 1
off = node_classification.get_off_nodes(conn, start, end)
offStorm = 0
for id in off:
    if id in storm:
        offStorm += 1

# Plot
plt.rc('font', size=8)
plt.pie([offStorm, onStorm, dangleStorm],
        labels=["off nodes %d" % offStorm, "on nodes %d" % onStorm, "dangling nodes %d" % dangleStorm],
        autopct="%.1f%%")
plt.title("Storm Nodes classification from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))
plt.show()

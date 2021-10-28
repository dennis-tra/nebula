import psycopg2
import toml
import matplotlib.pyplot as plt
from lib import node_time, node_classification, node_agent
import os
import psutil


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

# Helper function to trim agent ipfs version
def trim_agent_ipfs(agent):
    if agent.startswith("/"):
        agent = agent[1:]
    if agent.startswith("go-ipfs/0.7"):
        return "0.7.x"
    elif agent.startswith("go-ipfs/0.8"):
        return "0.8.x"
    elif agent.startswith("go-ipfs/0.9"):
        return "0.9.x"
    elif agent.startswith("go-ipfs/0.10"):
        return "0.10.x"
    elif agent.startswith("go-ipfs/0.11"):
        return "0.11.x"
    elif agent.startswith("go-ipfs"):
        return "older"
    else:
        return None

# Helper function to trim agent ipfs version 0.8.x
def trim_agent_ipfs_v08(version):
    if version.startswith("/"):
        version = version[1:]
    if not version.startswith("go-ipfs/0.8"):
        return None
    elif version == "go-ipfs/0.8.0/":
        return version[8:]
    elif version == "go-ipfs/0.8.0/16615d7":
        return version[8:]
    elif version == "go-ipfs/0.8.0/ce693d7":
        return version[8:]
    elif version == "go-ipfs/0.8.0/48f94e2":
        return version[8:]
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
all = node_classification.get_all_nodes(conn, start, end)
agents = node_agent.get_agent_version(conn, all)

# Plot
plt.rc('font', size=8)
fig, axs = plt.subplots(2, 2)
fig.delaxes(axs[1, 1])
fig.suptitle("All nodes agents from %s to %s" % (start.replace(microsecond=0), end.replace(microsecond=0)))

counts = dict()
for _, agent in agents.items():
    agent = trim_agent(agent)
    if agent in counts:
        counts[agent] += 1
    else:
        counts[agent] = 1
axs[0, 0].pie(counts.values(), labels=counts.keys(), autopct="%.1f%%")
axs[0, 0].set_title("Node agent version pie chart")

counts = dict()
for _, agent in agents.items():
    agent = trim_agent_ipfs(agent)
    if agent is None:
        continue
    if agent in counts:
        counts[agent] += 1
    else:
        counts[agent] = 1
axs[0, 1].pie(counts.values(), labels=counts.keys(), autopct="%.1f%%")
axs[0, 1].set_title("Node go-ipfs version pie chart")

counts = dict()
for _, agent in agents.items():
    agent = trim_agent_ipfs_v08(agent)
    if agent is None:
        continue
    if agent in counts:
        counts[agent] += 1
    else:
        counts[agent] = 1
axs[1, 0].pie(counts.values(), labels=counts.keys(), autopct="%.1f%%")
axs[1, 0].set_title("Node go-ipfs-0.8.x pie chart")

plt.savefig("./figs/agent_version_for_all_nodes.png")

print("memory used:", psutil.Process(os.getpid()).memory_info().rss / 1024 ** 2, "MB")
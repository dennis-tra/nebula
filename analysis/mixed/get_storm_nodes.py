import psycopg2
import toml
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

config = toml.load("./db.toml")['psql']
conn = psycopg2.connect(
    host=config['host'],
    port=config['port'],
    database=config['database'],
    user=config['user'],
    password=config['password'],
)
cur = conn.cursor()

cur.execute(
    """
    SELECT MAX(updated_at) - interval '1 hours', MAX(updated_at)
    FROM sessions
    """
)
record = cur.fetchone()
start, end = record[0].astimezone(), record[1].astimezone()
# Get storm node ids
all = node_classification.get_all_nodes(conn, start, end)
agents = node_agent.get_agent_version(conn, all)
storm = set()
for id, agent in agents.items():
    agent = trim_agent(agent)
    if agent == "storm":
        storm.add(id)

cur.execute(
    """
    SELECT id, multi_addresses, protocol
    FROM peers
    WHERE id IN (%s)
    """ % ','.join(['%s'] * len(storm)),
    tuple(storm)
)
with open("./storm.list", "w") as f:
    for id, maddr_strs, protocols in cur.fetchall():
        for maddr in maddr_strs:
            if maddr != "/p2p-circuit":
                f.write("{} {} {}\n".format(id, maddr, protocols))
                break
print("memory used:", psutil.Process(os.getpid()).memory_info().rss / 1024 ** 2, "MB")
from plot_agent import plot_agent
from lib_db import DBClient

client = DBClient()
peer_ids = client.get_online_peer_ids()
results = client.get_agent_versions_for_peer_ids(peer_ids)

plot_agent(results)

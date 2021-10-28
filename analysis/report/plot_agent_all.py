from lib_db import DBClient
from plot_agent import plot_agent

client = DBClient()
results = client.get_visited_peers_agent_versions()

plot_agent(results)

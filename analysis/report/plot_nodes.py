from collections import OrderedDict

import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
from matplotlib import ticker
from lib_agent import agent_name, known_agents, go_ipfs_version_mappings, go_ipfs_version
from lib_db import DBClient
from lib_fmt import fmt_thousands, fmt_barplot

sns.set_theme()

client = DBClient()

all_peer_ids = set(client.get_all_peer_ids())
online_peer_ids = set(client.get_online_peer_ids())
offline_peer_ids = set(client.get_offline_peer_ids())
all_entering_peer_ids = set(client.get_entering_peer_ids())
all_leaving_peer_ids = set(client.get_leaving_peer_ids())
dangling_peer_ids = all_entering_peer_ids.intersection(all_leaving_peer_ids)

entering_peer_ids = all_entering_peer_ids.difference(all_leaving_peer_ids)
leaving_peer_ids = all_leaving_peer_ids.difference(all_entering_peer_ids)

assert online_peer_ids.issubset(all_peer_ids)
assert offline_peer_ids.issubset(all_peer_ids)
assert entering_peer_ids.issubset(all_peer_ids)
assert leaving_peer_ids.issubset(all_peer_ids)
assert dangling_peer_ids.issubset(all_peer_ids)

assert online_peer_ids.isdisjoint(offline_peer_ids)
assert online_peer_ids.isdisjoint(entering_peer_ids)
assert online_peer_ids.isdisjoint(leaving_peer_ids)
assert online_peer_ids.isdisjoint(dangling_peer_ids)
assert offline_peer_ids.isdisjoint(entering_peer_ids)

assert offline_peer_ids.isdisjoint(leaving_peer_ids)
assert offline_peer_ids.isdisjoint(dangling_peer_ids)

assert entering_peer_ids.isdisjoint(leaving_peer_ids)
assert entering_peer_ids.isdisjoint(dangling_peer_ids)

assert leaving_peer_ids.isdisjoint(dangling_peer_ids)

data = OrderedDict([
    ("online", len(online_peer_ids)),
    ("offline", len(offline_peer_ids)),
    ("entered", len(entering_peer_ids)),
    ("left", len(leaving_peer_ids)),
    ("dangling", len(dangling_peer_ids)),
])
data = OrderedDict(reversed(sorted(data.items(), key=lambda item: item[1])))

# Plotting

fig, (ax) = plt.subplots()  # rows, cols

sns.barplot(ax=ax, x=list(data.keys()), y=list(data.values()))
fmt_barplot(ax, list(data.values()), len(all_peer_ids))

ax.title.set_text(f"Node Classification of {fmt_thousands(len(all_peer_ids))} Visited Peers")
ax.set_ylabel("Count")

plt.tight_layout()
plt.show()

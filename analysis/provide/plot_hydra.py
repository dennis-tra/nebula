import json
from models import *
import matplotlib.pyplot as plt

# Load measurement information
prefixes: list[str]
with open(f"server/measurements.json") as f:
    prefixes = json.load(f)

all_peers = 0
hydra_peers = 0
seen = {}

for prefix in prefixes:

    peer_infos: dict[str, PeerInfo] = {}
    with open(path.join("server", f"{prefix}_peer_infos.json")) as f:
        data = json.load(f)
        for key in data:
            peer_infos[key] = PeerInfo.from_dict(data[key])

    for peer_info in peer_infos.values():
        if peer_info.id in seen:
            continue
        seen[peer_info.id] = True
        all_peers += 1
        if "hydra" in peer_info.agent_version:
            hydra_peers += 1

print(hydra_peers)
print(all_peers)
print(hydra_peers / all_peers * 100, "%")

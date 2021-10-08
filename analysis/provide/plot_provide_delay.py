import json
import numpy as np
from models import *
import matplotlib.pyplot as plt

# Load measurement information
prefixes: list[str]
with open(f"data/measurements.json") as f:
    prefixes = json.load(f)

discovered_times = []

for prefix in prefixes:

    measurement = Measurement.from_location("data", prefix)

    selected_peers = []
    provided_times = []

    first_provide = None
    for spans in measurement.provider_spans.values():
        for span in spans:
            if span.type != "send_message":
                continue
            selected_peers += [span.peer_id]
            provided_times += [span.rel_start]

    for idx, peer in enumerate(selected_peers):
        provided_time = provided_times[idx]
        peer_info = measurement.peer_infos[peer]
        if peer_info.discovered_from == "":
            discovered_times += [provided_time]
        else:
            discovered_times += [provided_time - peer_info.rel_discovered_at]

plt.hist(discovered_times, bins=np.arange(100))


plt.xlabel("Discover Provide Delay in s")
plt.ylabel("Count")

plt.show()

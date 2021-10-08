import json
import numpy as np
from models import *
import matplotlib.pyplot as plt

# Load measurement information
prefixes: list[str]
with open(f"data/measurements.json") as f:
    prefixes = json.load(f)

distances = []
provide_times = []

for prefix in prefixes:

    measurement = Measurement.from_location("data", prefix)
    for spans in measurement.provider_spans.values():
        for span in spans:
            if span.type != "send_message":
                continue

            distances += [int(measurement.peer_infos[span.peer_id].xor_distance, base=16) / (2 ** 256) * 100]

plt.hist(distances, bins=np.arange(50)/100)

plt.title("Selected Peers by XOR Target Distance")
plt.ylabel("Count")
plt.xlabel("Normed XOR Distance")
plt.show()

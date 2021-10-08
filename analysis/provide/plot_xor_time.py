import json
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

    first_provide = None
    for spans in measurement.provider_spans.values():
        for span in spans:
            if span.type != "send_message":
                continue

            if first_provide is None or span.rel_start < first_provide:
                first_provide = span.rel_start

    provide_times += [first_provide]
    distances += [int(measurement.info.provider_dist, base=16) / (2 ** 256) * 100]


plt.ylabel("Time in s")

plt.xlim(0, 100)
plt.xlabel("Normed XOR Distance")

plt.scatter(distances, provide_times)
plt.show()

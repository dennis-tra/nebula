import json
import numpy as np
from models import *
import matplotlib.pyplot as plt

# Load measurement information
measurements: list[str]
with open(f"measurements.json") as f:
    measurements = json.load(f)

discovered_times = []

for measurement in measurements:

    # Load measurement information
    measurement_info: MeasurementInfo
    with open(f"{measurement}_measurement_info.json") as f:
        measurement_info = MeasurementInfo.from_dict(json.load(f))

    # Load peer information
    peer_infos: dict[str, PeerInfo] = {}
    with open(f"{measurement}_peer_infos.json") as f:
        data = json.load(f)
        for key in data:
            peer_infos[key] = PeerInfo.from_dict(data[key])

    # Load provider spans
    provider_spans: dict[str, list[Span]] = {}
    with open(f"{measurement}_provider_spans.json") as f:
        data = json.load(f)
        for key in data:
            provider_spans[key] = []
            for span_dict in data[key]:
                provider_spans[key] += [Span.from_dict(span_dict)]

    # Load requester spans
    requester_spans: dict[str, list[Span]] = {}
    with open(f"{measurement}_requester_spans.json") as f:
        data = json.load(f)
        for key in data:
            requester_spans[key] = []
            for span_dict in data[key]:
                requester_spans[key] += [Span.from_dict(span_dict)]

    selected_peers = []
    provided_times = []

    first_provide = None
    for spans in provider_spans.values():
        for span in spans:
            if span.type != "send_message":
                continue
            selected_peers += [span.peer_id]
            provided_times += [span.rel_start]

    for idx, peer in enumerate(selected_peers):
        provided_time = provided_times[idx]
        peer_info = peer_infos[peer]
        if peer_info.discovered_from == "":
            discovered_times += [provided_time]
        else:
            discovered_times += [provided_time - peer_info.rel_discovered_at]

plt.hist(discovered_times, bins=np.arange(100))


plt.xlabel("Discover Provide Delay in s")
plt.ylabel("Count")

plt.show()

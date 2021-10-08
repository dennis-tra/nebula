import json
from models import *
import matplotlib.pyplot as plt

# Load measurement information
measurements: list[str]
with open(f"measurements.json") as f:
    measurements = json.load(f)


distances = []
provide_times = []

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

    first_provide = None
    for spans in provider_spans.values():
        for span in spans:
            if span.type != "send_message":
                continue

            if first_provide is None or span.rel_start < first_provide:
                first_provide = span.rel_start

    provide_times += [first_provide]
    distances += [int(measurement_info.provider_dist, base=16) / (2 ** 256) * 100]


plt.ylabel("Time in s")

plt.xlim(0, 100)
plt.xlabel("Normed XOR Distance")

plt.scatter(distances, provide_times)
plt.show()

import json
import numpy as np
from models import *
import matplotlib.pyplot as plt

# Load measurement information
measurements: list[str]
with open(f"measurements.json") as f:
    measurements = json.load(f)

discovered_times = []

successful_dials = 0
failed_dials = 0

successful_queries = 0
failed_queries = 0

successful_provides = 0
failed_provides = 0

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

    for spans in provider_spans.values():
        for span in spans:
            if span.type == "dial":
                if span.error == "":
                    successful_dials += 1
                else:
                    failed_dials += 1
            elif span.type == "send_request":
                if span.error == "":
                    successful_queries += 1
                else:
                    failed_queries += 1
            elif span.type == "send_message":
                if span.error == "":
                    successful_provides += 1
                else:
                    failed_provides += 1

success_percent = successful_dials/(successful_dials+failed_dials)*100
print("Dials succeed by {:.1f}%".format(success_percent))

success_percent = successful_queries/(successful_queries+failed_queries)*100
print("Queries succeed by {:.1f}%".format(success_percent))

success_percent = successful_provides/(successful_provides+failed_provides)*100
print("Provides succeed by {:.1f}%".format(success_percent))

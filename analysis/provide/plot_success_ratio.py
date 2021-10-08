import json
import numpy as np
from models import *
import matplotlib.pyplot as plt

# Load measurement information
prefixes: list[str]
with open("./data/measurements.json") as f:
    prefixes = json.load(f)

discovered_times = []

successful_dials = 0
failed_dials = 0

successful_queries = 0
failed_queries = 0

successful_provides = 0
failed_provides = 0

for prefix in prefixes:

    measurement = Measurement.from_location("data", prefix)

    for spans in measurement.provider_spans.values():
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

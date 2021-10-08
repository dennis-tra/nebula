import json
import seaborn as sns
import matplotlib.pyplot as plt
import matplotlib.patches as mpatches
from models import *

sns.set_theme(style="darkgrid")
run_prefix = "2021-10-07T21:11:00"

span_colors = {
    "dial": sns.color_palette()[3],
    "send_request": sns.color_palette()[2],
    "send_message": sns.color_palette()[1],
}
span_colors_muted = {
    "dial": "#e8b0b0",
    "send_request": "#badec4",
    "send_message": "#f0c5a8",
}

colors = {
    'blue': sns.color_palette()[0],
    'orange': sns.color_palette()[1],
    'green': sns.color_palette()[2],
    'red': sns.color_palette()[3],
    'purple': sns.color_palette()[4],
    'brown': sns.color_palette()[5],
}

# Load measurement information
measurement_info: MeasurementInfo
with open(f"{run_prefix}_measurement_info.json") as f:
    measurement_info = MeasurementInfo.from_dict(json.load(f))

# Load peer information
peer_infos: dict[str, PeerInfo] = {}
with open(f"{run_prefix}_peer_infos.json") as f:
    data = json.load(f)
    for key in data:
        peer_infos[key] = PeerInfo.from_dict(data[key])

# Load provider spans
provider_spans: dict[str, list[Span]] = {}
with open(f"{run_prefix}_provider_spans.json") as f:
    data = json.load(f)
    for key in data:
        provider_spans[key] = []
        for span_dict in data[key]:
            provider_spans[key] += [Span.from_dict(span_dict)]

# Load requester spans
requester_spans: dict[str, list[Span]] = {}
with open(f"{run_prefix}_requester_spans.json") as f:
    data = json.load(f)
    for key in data:
        requester_spans[key] = []
        for span_dict in data[key]:
            requester_spans[key] += [Span.from_dict(span_dict)]

# Start plotting
num_peers = len(measurement_info.peer_order)

fig, ax = plt.subplots(1, figsize=(16, 6))
for peer_id in provider_spans:
    spans = provider_spans[peer_id]
    for span in spans:
        y = num_peers - measurement_info.peer_order.index(peer_id)
        xmin = span.rel_start
        xmax = span.rel_start + span.duration_s
        c = span_colors if span.error == "" else span_colors_muted
        ax.hlines(y, xmin, xmax, color=c[span.type], linewidth=3)
        if span.error != "":
            ax.plot([xmax], [y], marker='x', color=c[span.type], markersize=4)
        if span.type == 'send_message':
            ax.plot([xmin], [y], marker='.', color=c[span.type], markersize=4)

for peer_id in peer_infos:
    peer_info = peer_infos[peer_id]
    if peer_info.discovered_from == "":
        continue
    top = num_peers - measurement_info.peer_order.index(peer_info.discovered_from)
    bottom = num_peers - measurement_info.peer_order.index(peer_info.id)
    ax.arrow(
        peer_info.rel_discovered_at,
        top,
        0,
        bottom - top,
        head_width=0.01,
        head_length=0.5,
        color="#ccc",
        length_includes_head=True
    )

ax.vlines(0, 0, num_peers, linewidth=0.5, colors=colors['brown'])

labels = []
for peer_id in measurement_info.peer_order:
    distance = int(peer_infos[peer_id].xor_distance, base=16)
    distance_norm = distance / (2 ** 256)
    labels += ["{:s} | {:.2e} | {:s}".format(peer_infos[peer_id].agent_version, distance_norm, peer_id[:16])]

ax.set_yticklabels(
    labels,
    fontsize=8,
    fontname='Monospace',
    fontweight="bold")
ax.set_yticks(range(1, num_peers + 1))

plt.legend(handles=[
    mpatches.Patch(color=span_colors["dial"], label='Dial'),
    mpatches.Patch(color=span_colors["send_request"], label='Find Node'),
    mpatches.Patch(color=span_colors["send_message"], label='Add Provider'),
    mpatches.Patch(color="#ccc", label='Discovered'),
])

plt.xlabel("Time in s")
plt.xlim(0)
plt.tight_layout()
plt.show()

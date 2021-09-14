import psycopg2
import toml
import matplotlib.pyplot as plot
import time_check
from datetime import datetime
import node_classification
import cloud_check
import geolocation_check
import latency_calculation
import node_reliability
import dangling_filter

plot.rc('font', size=8)

config = toml.load("./db.toml")['psql']
conn = psycopg2.connect(
    host=config['host'],
    port=config['port'],
    database=config['database'],
    user=config['user'],
    password=config['password'],
)

# Get the start, end time
start, end = time_check.get_time_range(conn)

# Get nodes
off_nodes, on_nodes, dangling_nodes = node_classification.get_nodes(conn, start, end)

# Generate charts for all nodes
fig, axs = plot.subplots(4, 4)
fig.suptitle("All nodes from %s to %s" % (start, end))

axs[0, 0].pie([len(off_nodes), len(on_nodes), len(dangling_nodes)],
              labels=["off nodes", "on nodes", "dangling nodes"],
              autopct="%.1f%%")
axs[0, 0].set_title("Node classification pie chart")

axs[0, 1].bar(["off nodes", "on nodes", "dangling nodes"],
              [len(off_nodes), len(on_nodes), len(dangling_nodes)])
axs[0, 1].set_title("Node classification bar chart")

axs[0, 2].set_title("Agent version pie chart")
axs[0, 3].set_title("Agent version bar chart")

axs[1, 0].set_title("Known agent version pie chart")
axs[1, 1].set_title("Known agent version bar chart")

axs[1, 2].set_title("Go-ipfs version pie chart")
axs[1, 3].set_title("Go-ipfs version bar chart")

axs[2, 0].set_title("Go-ipfs 0.8.x version pie chart")
axs[2, 1].set_title("Go-ipfs 0.8.x version bar chart")

axs[2, 2].set_title("Agent protocols pie chart")
axs[2, 3].set_title("Agent protocols bar chart")

cloud_info = cloud_check.check_cloud(conn, off_nodes + on_nodes + dangling_nodes)
counts = dict()
for _, provider in cloud_info.items():
    if provider in counts:
        counts[provider] += 1
    else:
        counts[provider] = 1
axs[3, 0].pie(counts.values(), labels=counts.keys(), autopct="%.1f%%")
axs[3, 0].set_title("Cloud info pie chart")
axs[3, 1].bar(counts.keys(), counts.values())
axs[3, 1].set_title("Cloud info bar chart")

location_info = geolocation_check.check_geolocation(conn, off_nodes + on_nodes + dangling_nodes)
counts = dict()
for _, location in location_info.items():
    if location in counts:
        counts[location] += 1
    else:
        counts[location] = 1
# Manipulate
countsFix = dict()
countsFix['others'] = 0
for key, val in counts.items():
    if val < 300:
        countsFix['others'] += val
    else:
        countsFix[key] = val
countsFix = dict(sorted(countsFix.items(), key=lambda item: item[1], reverse=True))

axs[3, 2].pie(countsFix.values(), labels=countsFix.keys(), autopct="%.1f%%")
axs[3, 2].set_title("Location info pie chart")
axs[3, 3].bar(countsFix.keys(), countsFix.values())
axs[3, 3].set_title("Location info bar chart")

# Generate charts for all on nodes
fig, axs = plot.subplots(4, 4)
fig.suptitle("On nodes from %s to %s" % (start, end))

latencies = latency_calculation.get_latency(conn, start, end, on_nodes)
avgs = []
maxes = []
for _, latency in latencies.items():
    avgs.append(latency[2].microseconds / 1000)
    maxes.append(latency[0].microseconds / 1000)
axs[0, 0].hist(avgs, density=False, bins=50)
axs[0, 0].set_title("Node average latency histogram (x-ms y-count)")

axs[0, 1].hist(maxes, density=False, bins=50)
axs[0, 1].set_title("Node maximum latency histogram (x-ms y-count)")

axs[0, 2].set_title("Agent version pie chart")
axs[0, 3].set_title("Agent version bar chart")

axs[1, 0].set_title("Known agent version pie chart")
axs[1, 1].set_title("Known agent version bar chart")

axs[1, 2].set_title("Go-ipfs version pie chart")
axs[1, 3].set_title("Go-ipfs version bar chart")

axs[2, 0].set_title("Go-ipfs 0.8.x version pie chart")
axs[2, 1].set_title("Go-ipfs 0.8.x version bar chart")

axs[2, 2].set_title("Agent protocols pie chart")
axs[2, 3].set_title("Agent protocols bar chart")

cloud_info = cloud_check.check_cloud(conn, on_nodes)
counts = dict()
for _, provider in cloud_info.items():
    if provider in counts:
        counts[provider] += 1
    else:
        counts[provider] = 1
axs[3, 0].pie(counts.values(), labels=counts.keys(), autopct="%.1f%%")
axs[3, 0].set_title("Cloud info pie chart")
axs[3, 1].bar(counts.keys(), counts.values())
axs[3, 1].set_title("Cloud info bar chart")

location_info = geolocation_check.check_geolocation(conn, on_nodes)
counts = dict()
for _, location in location_info.items():
    if location in counts:
        counts[location] += 1
    else:
        counts[location] = 1
# Manipulate
countsFix = dict()
countsFix['others'] = 0
for key, val in counts.items():
    if val < 100:
        countsFix['others'] += val
    else:
        countsFix[key] = val
countsFix = dict(sorted(countsFix.items(), key=lambda item: item[1], reverse=True))

axs[3, 2].pie(countsFix.values(), labels=countsFix.keys(), autopct="%.1f%%")
axs[3, 2].set_title("Location info pie chart")
axs[3, 3].bar(countsFix.keys(), countsFix.values())
axs[3, 3].set_title("Location info bar chart")

# Generate charts for all dangling nodes
fig, axs = plot.subplots(4, 4)
fig.suptitle("Dangling nodes from %s to %s" % (start, end))

latencies = latency_calculation.get_latency(conn, start, end, dangling_nodes)
avgs = []
maxes = []
for _, latency in latencies.items():
    avgs.append(latency[2].microseconds / 1000)
    maxes.append(latency[0].microseconds / 1000)
axs[0, 0].hist(avgs, density=False, bins=50)
axs[0, 0].set_title("Node average latency histogram (x-ms y-count)")

axs[0, 1].hist(maxes, density=False, bins=50)
axs[0, 1].set_title("Node maximum latency histogram (x-ms y-count)")

reliabilities = node_reliability.get_reliability(conn, start, end, dangling_nodes)
axs[0, 2].hist(reliabilities, density=False, bins=50)
axs[0, 2].set_title("Node reliabilities")

switches = dangling_filter.get_dangling_counts(conn, start, end, dangling_nodes)
axs[0, 3].hist(switches.values(), density=False, bins=50)
axs[0, 3].set_title("Node switches")

axs[1, 0].set_title("Known agent version pie chart")
axs[1, 1].set_title("Known agent version bar chart")

axs[1, 2].set_title("Go-ipfs version pie chart")
axs[1, 3].set_title("Go-ipfs version bar chart")

axs[2, 0].set_title("Go-ipfs 0.8.x version pie chart")
axs[2, 1].set_title("Go-ipfs 0.8.x version bar chart")

axs[2, 2].set_title("Agent protocols pie chart")
axs[2, 3].set_title("Agent protocols bar chart")

cloud_info = cloud_check.check_cloud(conn, dangling_nodes)
counts = dict()
for _, provider in cloud_info.items():
    if provider in counts:
        counts[provider] += 1
    else:
        counts[provider] = 1
axs[3, 0].pie(counts.values(), labels=counts.keys(), autopct="%.1f%%")
axs[3, 0].set_title("Cloud info pie chart")
axs[3, 1].bar(counts.keys(), counts.values())
axs[3, 1].set_title("Cloud info bar chart")

location_info = geolocation_check.check_geolocation(conn, dangling_nodes)
counts = dict()
for _, location in location_info.items():
    if location in counts:
        counts[location] += 1
    else:
        counts[location] = 1
# Manipulate
countsFix = dict()
countsFix['others'] = 0
for key, val in counts.items():
    if val < 300:
        countsFix['others'] += val
    else:
        countsFix[key] = val
countsFix = dict(sorted(countsFix.items(), key=lambda item: item[1], reverse=True))

axs[3, 2].pie(countsFix.values(), labels=countsFix.keys(), autopct="%.1f%%")
axs[3, 2].set_title("Location info pie chart")
axs[3, 3].bar(countsFix.keys(), countsFix.values())
axs[3, 3].set_title("Location info bar chart")

# Highly dangling nodes
highly_dangling_nodes = []
for key, val in switches.items():
    if val >= 2:
        highly_dangling_nodes.append(key)

fig, axs = plot.subplots(4, 4)
fig.suptitle("Highly Dangling nodes from %s to %s" % (start, end))

latencies = latency_calculation.get_latency(conn, start, end, highly_dangling_nodes)
avgs = []
maxes = []
for _, latency in latencies.items():
    avgs.append(latency[2].microseconds / 1000)
    maxes.append(latency[0].microseconds / 1000)
axs[0, 0].hist(avgs, density=False, bins=50)
axs[0, 0].set_title("Node average latency histogram (x-ms y-count)")

axs[0, 1].hist(maxes, density=False, bins=50)
axs[0, 1].set_title("Node maximum latency histogram (x-ms y-count)")

reliabilities = node_reliability.get_reliability(conn, start, end, highly_dangling_nodes)
axs[0, 2].hist(reliabilities, density=False, bins=50)
axs[0, 2].set_title("Node reliabilities")

switches = dangling_filter.get_dangling_counts(conn, start, end, highly_dangling_nodes)
axs[0, 3].hist(switches.values(), density=False, bins=50)
axs[0, 3].set_title("Node switches")

axs[1, 0].set_title("Known agent version pie chart")
axs[1, 1].set_title("Known agent version bar chart")

axs[1, 2].set_title("Go-ipfs version pie chart")
axs[1, 3].set_title("Go-ipfs version bar chart")

axs[2, 0].set_title("Go-ipfs 0.8.x version pie chart")
axs[2, 1].set_title("Go-ipfs 0.8.x version bar chart")

axs[2, 2].set_title("Agent protocols pie chart")
axs[2, 3].set_title("Agent protocols bar chart")

cloud_info = cloud_check.check_cloud(conn, dangling_nodes)
counts = dict()
for _, provider in cloud_info.items():
    if provider in counts:
        counts[provider] += 1
    else:
        counts[provider] = 1
axs[3, 0].pie(counts.values(), labels=counts.keys(), autopct="%.1f%%")
axs[3, 0].set_title("Cloud info pie chart")
axs[3, 1].bar(counts.keys(), counts.values())
axs[3, 1].set_title("Cloud info bar chart")

location_info = geolocation_check.check_geolocation(conn, highly_dangling_nodes)
counts = dict()
for _, location in location_info.items():
    if location in counts:
        counts[location] += 1
    else:
        counts[location] = 1
# Manipulate
countsFix = dict()
countsFix['others'] = 0
for key, val in counts.items():
    if val < 100:
        countsFix['others'] += val
    else:
        countsFix[key] = val
countsFix = dict(sorted(countsFix.items(), key=lambda item: item[1], reverse=True))

axs[3, 2].pie(countsFix.values(), labels=countsFix.keys(), autopct="%.1f%%")
axs[3, 2].set_title("Location info pie chart")
axs[3, 3].bar(countsFix.keys(), countsFix.values())
axs[3, 3].set_title("Location info bar chart")

# Show all graphs
plot.show()

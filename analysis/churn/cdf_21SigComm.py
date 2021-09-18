import csv
import matplotlib.pyplot as plt
import numpy as np
import matplotlib
from matplotlib.pyplot import figure
plt.rcParams.update({
    "font.family": "serif",
    "font.serif": [],
})
matplotlib.rcParams['text.usetex'] = True

latex_line_width_pt = 241.14749
latex_line_width_in = latex_line_width_pt / 72.27


def format_interval(x, pos=None):
    return int(x / 3600)


ipfs_min_durs = []
with open("sessions.csv") as csvfile:
    csvreader = csv.DictReader(csvfile, delimiter=",")

    for row in csvreader:
        # If the session is not finished yet -> skip
        if row["min_duration_s"] == "":
            continue

        ipfs_min_durs += [float(row["min_duration_s"])]

filecoin_min_durs = []
with open("sessions_filecoin.csv") as csvfile:
    csvreader = csv.DictReader(csvfile, delimiter=",")

    for row in csvreader:
        # If the session is not finished yet -> skip
        if row["min_duration_s"] == "":
            continue

        filecoin_min_durs += [float(row["min_duration_s"])]



#### ALL IPFS ####

# Creates a histogram of the duration of sessions
hist_values, bin_edges = np.histogram(
    ipfs_min_durs, bins=len(ipfs_min_durs), density=True
)

# Since we provided an integer to the bins parameter above. The edges are equal width.
# This means the width between the first two elements is equal for all edges.
edge_width = bin_edges[1] - bin_edges[0]

# Integerate over histogram
cumsum = np.cumsum(hist_values) * edge_width

# build plot
plt.plot(bin_edges[1:], cumsum*100, label="IPFS")

#### ALL FILECOIN ####

# Creates a histogram of the duration of sessions
hist_values, bin_edges = np.histogram(
    filecoin_min_durs, bins=len(filecoin_min_durs), density=True
)

# Since we provided an integer to the bins parameter above. The edges are equal width.
# This means the width between the first two elements is equal for all edges.
edge_width = bin_edges[1] - bin_edges[0]

# Integerate over histogram
cumsum = np.cumsum(hist_values) * edge_width

# build plot
plt.plot(bin_edges[1:], cumsum*100, label="Filecoin")

#############

# Presentation logic
plt.gca().xaxis.set_major_formatter(format_interval)
# plt.xticks(rotation=45, ha="right")
plt.xlabel(r"Time in Hours")
plt.ylabel(r"Online Peers in \%")
plt.xticks(np.arange(0, max(bin_edges[1:]), 3 * 60 * 60))
plt.yticks(np.arange(0, 110, 20))
plt.xlim(-60 * 60, 24 * 60 * 60)

plt.grid(True)
plt.legend()

# Finalize
# plt.show()
fig = plt.gcf()
fig.set_size_inches(latex_line_width_in, latex_line_width_in / 1.5)
plt.tight_layout()
plt.savefig("cdf_21SigCommm.pdf")
plt.savefig("cdf_21SigCommm.pgf", backend='pgf')

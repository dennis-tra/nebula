import csv
import matplotlib.pyplot as plt
import numpy as np


def format_interval(x, pos=None):
    return x / 3600


with open("sessions_filecoin.csv") as csvfile:
    csvreader = csv.DictReader(csvfile, delimiter=",")

    all_min_durs = []
    for row in csvreader:

        # If the session is not finished yet -> skip
        if row["min_duration_s"] == "":
            all_min_durs += [6 * 24 * 60 * 60]
            continue

        all_min_durs += [float(row["min_duration_s"])]

    #### ALL ####

    # Creates a histogram of the duration of sessions
    hist_values, bin_edges = np.histogram(
        all_min_durs, bins=len(all_min_durs), density=True
    )

    # Since we provided an integer to the bins parameter above. The edges are equal width.
    # This means the width between the first two elements is equal for all edges.
    edge_width = bin_edges[1] - bin_edges[0]

    # Integerate over histogram
    cumsum = np.cumsum(hist_values) * edge_width

    # build plot
    plt.plot(bin_edges[1:], cumsum)

    #############

    # Presentation logic
    plt.gca().xaxis.set_major_formatter(format_interval)
    # plt.xticks(rotation=45, ha="right")
    plt.xlabel("Hours")
    plt.ylabel("Percentage of online peers")
    plt.xticks(np.arange(0, max(bin_edges[1:]), 12 * 60 * 60))
    plt.xlim(-60 * 60, 5 * 24 * 60 * 60)

    plt.title(f"Based on {len(all_min_durs)} observed sessions")
    plt.grid(True)
    # plt.legend()
    plt.tight_layout()

    # Finalize
    # plt.show()
    plt.savefig("filecoin_cdf.png")

import csv
import matplotlib.pyplot as plt
import numpy as np
import datetime


def format_interval(x, pos=None):
    return x / 3600


deny_list = [
    "138.68.226.34",
    "138.68.226.34",
    "138.68.31.185",
    "165.227.61.39",
    "138.68.230.227",
    "138.197.203.17",
    "138.68.227.119",
    "165.227.25.67",
    "165.227.1.120",
    "138.68.0.5",
    "138.68.23.43",
    "138.68.0.5",
    "138.68.23.43",
]


with open("sessions.csv") as csvfile:
    csvreader = csv.DictReader(csvfile, delimiter=",")

    skipped_sessions = 0
    all_min_durs = []
    filtered_min_durs = []
    for row in csvreader:

        # If the session is not finished yet -> skip
        if row["min_duration_s"] == "":
            # filtered_min_durs += [5 * 24 * 60 * 60]
            continue

        all_min_durs += [float(row["min_duration_s"])]

        # Check if ip is in deny list
        is_in_deny_list = False
        for ip in deny_list:
            if ip in row["multi_addresses"]:
                is_in_deny_list = True
                skipped_sessions += 1
                break
        if is_in_deny_list:
            continue

        filtered_min_durs += [float(row["min_duration_s"])]

    print("Skipped sessions: " + str(skipped_sessions))

    #### FILTERED ####

    # Creates a histogram of the duration of sessions
    hist_values, bin_edges = np.histogram(
        filtered_min_durs, bins=len(filtered_min_durs), density=True
    )

    # Since we provided an integer to the bins parameter above. The edges are equal width.
    # This means the width between the first two elements is equal for all edges.
    edge_width = bin_edges[1] - bin_edges[0]

    # Integerate over histogram
    cumsum = np.cumsum(hist_values) * edge_width

    # build plot
    plt.plot(bin_edges[1:], cumsum, label="Filtered sessions")

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
    plt.plot(bin_edges[1:], cumsum, label="All sessions")

    #############

    # Presentation logic
    plt.gca().xaxis.set_major_formatter(format_interval)
    # plt.xticks(rotation=45, ha="right")
    plt.xlabel("Hours")
    plt.ylabel("Percentage of online peers")
    plt.tight_layout()
    plt.xticks(np.arange(0, max(bin_edges[1:]), 3 * 60 * 60))
    plt.xlim(-60 * 60, 24 * 60 * 60)

    plt.grid(True)
    plt.legend()

    # Finalize
    plt.show()

import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

import lib_plot
from lib_db import DBClient
from lib_fmt import fmt_thousands, fmt_barplot

sns.set_theme()

client = DBClient()


def plot_geo(peer_ids, classification, threshold, file_name):
    results = client.get_country_distribution_for_peer_ids(peer_ids)
    data = pd.DataFrame(results, columns=["Country", "Count"])

    fig, ax = plt.subplots(figsize=(12, 6))

    # calculate the "other" countries
    granular_df = data[data["Count"] > threshold]
    others_df = data[data["Count"] <= threshold]
    others_sum_df = pd.DataFrame([["other", others_df["Count"].sum()]], columns=["Country", "Count"])
    all_df = granular_df.append(others_sum_df)

    sns.barplot(ax=ax, x="Country", y="Count", data=all_df)
    fmt_barplot(ax, all_df["Count"], all_df["Count"].sum())

    plt.title(f"Country Distribution of {classification} Peers (Total {fmt_thousands(data['Count'].sum())})")

    lib_plot.savefig(f"geo-{file_name}")
    plt.show()


peer_ids = client.get_dangling_peer_ids()
plot_geo(peer_ids, "Dangling", 200, "dangling")

peer_ids = client.get_offline_peer_ids()
plot_geo(peer_ids, "Offline", 200, "offline")

peer_ids = client.get_online_peer_ids()
plot_geo(peer_ids, "Online", 15, "online")

peer_ids = client.get_peer_ids_for_agent_versions(["hydra-booster/0.7.4"])
plot_geo(peer_ids, "'hydra-booster/0.7.4'", 15, "hydra")

peer_ids = client.get_peer_ids_for_agent_versions(["ioi"])
plot_geo(peer_ids, "'ioi'", 20, "ioi")

peer_ids = client.get_peer_ids_for_agent_versions(["storm"])
plot_geo(peer_ids, "'storm'", 15, "storm")

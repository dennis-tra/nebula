import seaborn as sns
import pandas as pd
import lib_plot
import matplotlib.pyplot as plt

from lib_agent import known_agents
from lib_db import DBClient
from lib_fmt import fmt_thousands, fmt_barplot


def main(db_client: DBClient):
    sns.set_theme()

    country_distributions = {}
    thresholds = {
        "go-ipfs": 350,
        "hydra-booster": 0,
        "storm": 50,
        "ioi": 10
    }

    for agent in known_agents:
        peer_ids = set(db_client.get_peer_ids_for_agent_versions([agent]))
        country_distributions[agent] = db_client.get_country_distribution_for_peer_ids(peer_ids)

    fig, axs = plt.subplots(2, 2, figsize=(15, 9))
    for idx, agent in enumerate(country_distributions):
        data = country_distributions[agent]
        ax = axs[idx // 2][idx % 2]

        # calculate the "other" countries
        granular_df = data[data["Count"] > thresholds[agent]]
        others_df = data[data["Count"] <= thresholds[agent]]
        others_sum_df = pd.DataFrame([["other", others_df["Count"].sum()]], columns=["Country", "Count"])
        all_df = granular_df.append(others_sum_df)

        sns.barplot(ax=ax, x="Country", y="Count", data=all_df)
        fmt_barplot(ax, all_df["Count"], all_df["Count"].sum())
        ax.set_xlabel("")
        ax.title.set_text(f"{agent} (Total {fmt_thousands(data['Count'].sum())})")

    plt.suptitle(f"Country Distributions of all Resolved Peer IDs by Agent Version")
    plt.tight_layout()
    lib_plot.savefig(f"geo-agents")
    plt.show()


if __name__ == '__main__':
    db_client = DBClient()
    main(db_client)

import seaborn as sns
import pandas as pd
from lib import lib_plot
import matplotlib.pyplot as plt

from lib.lib_db_filecoin import DBClientFilecoin
from lib.lib_db import DBClient
from lib.lib_fmt import fmt_thousands, fmt_barplot


def main(db_client: DBClient):
    sns.set_theme()

    country_distributions = {}

    results = db_client.get_agent_versions_distribution()
    top_agent_versions = [result[0] for result in results[:4]]

    for agent in top_agent_versions:
        peer_ids = set(db_client.get_peer_ids_for_agent_versions([agent]))
        country_distributions[agent] = db_client.get_country_distribution_for_peer_ids(peer_ids)

    fig, axs = plt.subplots(2, 2, figsize=(15, 9))
    for idx, agent in enumerate(country_distributions):
        data = country_distributions[agent]
        ax = axs[idx // 2][idx % 2]

        # calculate the "other" countries
        granular_df = data[data["Count"] > 20]
        others_df = data[data["Count"] <= 20]
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
    db_client = DBClientFilecoin()
    main(db_client)

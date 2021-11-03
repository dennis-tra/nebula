import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

import lib_plot
from lib_db import DBClient, NodeClassification
from lib_fmt import fmt_thousands, fmt_barplot


def main(db_client: DBClient):
    sns.set_theme()

    country_distributions = {}
    thresholds_ipfs = {
        NodeClassification.OFFLINE: 300,
        NodeClassification.ONEOFF: 300,
        NodeClassification.DANGLING: 500,
        NodeClassification.ONLINE: 50,
        NodeClassification.ENTERED: 14,
        NodeClassification.LEFT: 15,
    }

    thresholds_filecoin = {
        NodeClassification.OFFLINE: 10,
        NodeClassification.ONEOFF: 5,
        NodeClassification.DANGLING: 12,
        NodeClassification.ONLINE: 10,
        NodeClassification.ENTERED: 5,
        NodeClassification.LEFT: 5,
    }

    thresholds = thresholds_ipfs

    for node_class in NodeClassification:
        peer_ids = db_client.node_classification_funcs[node_class]()
        country_distributions[node_class] = db_client.get_country_distribution_for_peer_ids(peer_ids)

    fig, axs = plt.subplots(2, 3, figsize=(15, 8))

    for idx, node_class in enumerate(country_distributions):
        data = country_distributions[node_class]
        ax = axs[idx // 3][idx % 3]

        # calculate the "other" countries
        granular_df = data[data["Count"] > thresholds[node_class]]
        others_df = data[data["Count"] <= thresholds[node_class]]
        others_sum_df = pd.DataFrame([["other", others_df["Count"].sum()]], columns=["Country", "Count"])
        all_df = granular_df.append(others_sum_df)

        sns.barplot(ax=ax, x="Country", y="Count", data=all_df)
        fmt_barplot(ax, all_df["Count"], all_df["Count"].sum())
        ax.set_xlabel("")

        ax.title.set_text(f"{node_class.value} (Total {fmt_thousands(data['Count'].sum())})")

    plt.suptitle(f"Country Distributions by Node Classification")
    plt.tight_layout()
    lib_plot.savefig(f"geo-node-classification")
    plt.show()


if __name__ == '__main__':
    db_client = DBClient()
    main(db_client)

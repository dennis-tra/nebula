import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

import lib_plot
from lib_db import DBClient, NodeClassification
from lib_cloud import Cloud
from lib_fmt import fmt_barplot, fmt_thousands


def main(db_client: DBClient, cloud_client: Cloud):
    sns.set_theme()

    ip_addresses = {}
    for node_class in NodeClassification:
        peer_ids = db_client.node_classification_funcs[node_class]()
        ip_addresses[node_class] = db_client.get_ip_addresses_for_peer_ids(peer_ids)

    fig, axs = plt.subplots(2, 3, figsize=(15, 8))

    for idx, node_class in enumerate(ip_addresses):
        data = ip_addresses[node_class]
        ax = axs[idx // 3][idx % 3]

        results_df = pd.DataFrame(data, columns=["ip_address"]).assign(
            cloud=lambda df: df.ip_address.apply(lambda ip: cloud_client.cloud_for(ip)),
            count=lambda df: df.ip_address.apply(lambda ip: 1),
        ).groupby(by='cloud', as_index=False).sum().sort_values('count', ascending=False)

        sns.barplot(ax=ax, x="cloud", y="count", data=results_df)
        fmt_barplot(ax, results_df["count"], results_df['count'].sum())

        ax.set_xlabel("")
        ax.set_ylabel("Count")

        ax.title.set_text(f"{node_class.value} (Total {fmt_thousands(results_df['count'].sum())})")

    plt.suptitle(f"Cloud Platform Distribution by Node Classification")

    plt.tight_layout()
    lib_plot.savefig(f"cloud-classification")
    plt.show()


if __name__ == '__main__':
    db_client = DBClient()
    cloud_client = Cloud()
    main(db_client, cloud_client)

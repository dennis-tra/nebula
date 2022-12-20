import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

from lib import lib_plot
from lib.lib_agent import known_agents
from lib.lib_db import DBClient
from lib.lib_cloud import Cloud
from lib.lib_fmt import fmt_barplot, fmt_thousands


def main(db_client: DBClient, cloud_client: Cloud):
    sns.set_theme()

    ip_addresses = {}
    for agent in known_agents:
        peer_ids = set(db_client.get_peer_ids_for_agent_versions([agent]))
        ip_addresses[agent] = db_client.get_ip_addresses_for_peer_ids(peer_ids)

    fig, axs = plt.subplots(2, 2, figsize=(15, 9))

    for idx, agent in enumerate(ip_addresses):
        data = ip_addresses[agent]
        ax = axs[idx // 2][idx % 2]

        results_df = pd.DataFrame(data, columns=["ip_address"]).assign(
            cloud=lambda df: df.ip_address.apply(lambda ip: cloud_client.cloud_for(ip)),
            count=lambda df: df.ip_address.apply(lambda ip: 1),
        ).groupby(by='cloud', as_index=False).sum().sort_values('count', ascending=False)

        sns.barplot(ax=ax, x="cloud", y="count", data=results_df)
        fmt_barplot(ax, results_df["count"], results_df['count'].sum())

        ax.set_xlabel("")
        ax.set_ylabel("Count")

        ax.title.set_text(f"{agent} (Total {fmt_thousands(results_df['count'].sum())})")

    plt.suptitle(f"Cloud Platform Distribution by Agent Version")

    plt.tight_layout()
    lib_plot.savefig(f"cloud-agents")
    plt.show()


if __name__ == '__main__':
    db_client = DBClient()
    cloud_client = Cloud()
    main(db_client, cloud_client)

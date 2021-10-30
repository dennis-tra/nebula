import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

import lib_plot
from lib_db import DBClient
from lib_cloud import Cloud
from lib_fmt import fmt_barplot, fmt_thousands

sns.set_theme()

client = DBClient()
cloud_client = Cloud()


def plot_cloud(data, classification, file_name):
    results_df = pd.DataFrame(data, columns=["ip_address"]).assign(
        cloud=lambda df: df.ip_address.apply(lambda ip: cloud_client.cloud_for(ip)),
        count=lambda df: df.ip_address.apply(lambda ip: 1),
    ).groupby(by='cloud', as_index=False).sum().sort_values('count', ascending=False)

    fig, ax = plt.subplots(figsize=(15, 5))

    sns.barplot(ax=ax, x="cloud", y="count", data=results_df)
    fmt_barplot(ax, results_df["count"], results_df['count'].sum())

    ax.set_xlabel("Cloud Platform")
    ax.set_ylabel("Count")

    plt.title(
        f"Cloud Platform Distribution of {classification} Peers (Total {fmt_thousands(results_df['count'].sum())})")

    plt.tight_layout()
    lib_plot.savefig(f"cloud-{file_name}")
    plt.show()


peer_ids = client.get_dangling_peer_ids()
results = client.get_ip_addresses_for_peer_ids(peer_ids)
plot_cloud(results, "Dangling", "dangling")

peer_ids = client.get_offline_peer_ids()
results = client.get_ip_addresses_for_peer_ids(peer_ids)
plot_cloud(results, "Offline", "offline")

peer_ids = client.get_online_peer_ids()
results = client.get_ip_addresses_for_peer_ids(peer_ids)
plot_cloud(results, "Online", "online")

peer_ids = client.get_peer_ids_for_agent_versions(["hydra-booster/0.7.4"])
results = client.get_ip_addresses_for_peer_ids(peer_ids)
plot_cloud(results, "'hydra-booster/0.7.4'", "hydra")

peer_ids = client.get_peer_ids_for_agent_versions(["storm"])
results = client.get_ip_addresses_for_peer_ids(peer_ids)
plot_cloud(results, "'storm'", "storm")

peer_ids = client.get_peer_ids_for_agent_versions(["ioi"])
results = client.get_ip_addresses_for_peer_ids(peer_ids)
plot_cloud(results, "'ioi'", "ioi")

results = client.query(
    """
    WITH cte AS (
        SELECT v.peer_id, unnest(mas.multi_address_ids) multi_address_id
        FROM visits v
                 INNER JOIN multi_addresses_sets mas on mas.id = v.multi_addresses_set_id
        WHERE v.created_at > date_trunc('week', NOW() - '1 week'::interval)
          AND v.created_at < date_trunc('week', NOW())
        GROUP BY v.peer_id, unnest(mas.multi_address_ids)
    )
    SELECT DISTINCT ia.address
    FROM multi_addresses ma
            INNER JOIN cte ON cte.multi_address_id = ma.id
            INNER JOIN multi_addresses_x_ip_addresses maxia on ma.id = maxia.multi_address_id
            INNER JOIN ip_addresses ia ON maxia.ip_address_id = ia.id
    """
)

plot_cloud(results, "All", "all")

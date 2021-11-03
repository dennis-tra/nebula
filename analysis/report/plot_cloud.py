import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

import lib_plot
from lib_db import DBClient
from lib_cloud import Cloud
from lib_fmt import fmt_barplot, fmt_thousands


def main(db_client: DBClient, cloud_client: Cloud):
    sns.set_theme()

    results = db_client.query(
        f"""
        WITH cte AS (
            SELECT v.peer_id, unnest(mas.multi_address_ids) multi_address_id
            FROM visits v
                     INNER JOIN multi_addresses_sets mas on mas.id = v.multi_addresses_set_id
            WHERE v.created_at > {db_client.start}
              AND v.created_at < {db_client.end}
            GROUP BY v.peer_id, unnest(mas.multi_address_ids)
        )
        SELECT DISTINCT ia.address
        FROM multi_addresses ma
                INNER JOIN cte ON cte.multi_address_id = ma.id
                INNER JOIN multi_addresses_x_ip_addresses maxia on ma.id = maxia.multi_address_id
                INNER JOIN ip_addresses ia ON maxia.ip_address_id = ia.id
        """
    )
    results_df = pd.DataFrame(results, columns=["ip_address"]).assign(
        cloud=lambda df: df.ip_address.apply(lambda ip: cloud_client.cloud_for(ip)),
        count=lambda df: df.ip_address.apply(lambda ip: 1),
    ).groupby(by='cloud', as_index=False).sum().sort_values('count', ascending=False)

    fig, ax = plt.subplots(figsize=(15, 5))

    sns.barplot(ax=ax, x="cloud", y="count", data=results_df)
    fmt_barplot(ax, results_df["count"], results_df['count'].sum())

    ax.set_xlabel("Cloud Platform")
    ax.set_ylabel("Count")

    plt.title(
        f"Cloud Platform Distribution of All Peers (Total {fmt_thousands(results_df['count'].sum())})")

    plt.tight_layout()
    lib_plot.savefig(f"cloud-all")
    plt.show()


if __name__ == '__main__':
    client = DBClient()
    cloud_client = Cloud()
    main(client, cloud_client)

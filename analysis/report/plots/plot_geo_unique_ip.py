import seaborn as sns
import pandas as pd
from matplotlib import pyplot as plt

from lib import lib_plot
from lib.lib_fmt import fmt_barplot, fmt_thousands
from lib.lib_db import DBClient


def main(db_client: DBClient, threshold=500):
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
        SELECT ia.country, count(DISTINCT ia.address) count
        FROM multi_addresses ma
                 INNER JOIN cte ON cte.multi_address_id = ma.id
                 INNER JOIN multi_addresses_x_ip_addresses maxia on ma.id = maxia.multi_address_id
                 INNER JOIN ip_addresses ia ON maxia.ip_address_id = ia.id
        GROUP BY ia.country
        ORDER BY count DESC
        """
    )

    data = pd.DataFrame(results, columns=["Country", "Count"])
    # calculate the "other" countries
    granular_df = data[data["Count"] > threshold]
    others_df = data[data["Count"] <= threshold]
    others_sum_df = pd.DataFrame([["other", others_df["Count"].sum()]], columns=["Country", "Count"])
    all_df = granular_df.append(others_sum_df)

    fig, ax = plt.subplots(figsize=(10, 5))

    sns.barplot(ax=ax, x="Country", y="Count", data=all_df)
    fmt_barplot(ax, all_df["Count"], all_df["Count"].sum())
    ax.set_xlabel("")

    plt.title(f"Country Distribution of Unique IP Addresses (Total {fmt_thousands(data['Count'].sum())})")

    plt.tight_layout()
    lib_plot.savefig(f"geo-unique-ip")
    plt.show()


if __name__ == '__main__':
    db_client = DBClient()
    main(db_client)

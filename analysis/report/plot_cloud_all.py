import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

from lib_cloud import Cloud
from lib_db import DBClient
from lib_fmt import fmt_barplot, fmt_thousands
from lib_agent import agent_name, go_ipfs_version, go_ipfs_v08_version

sns.set_theme()

client = DBClient()
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

cloud_client = Cloud()

results_df = pd.DataFrame(results, columns=["ip_address"]).assign(
    cloud=lambda df: df.ip_address.apply(lambda ip: cloud_client.cloud_for(ip)),
    count=lambda df: df.ip_address.apply(lambda ip: 1),
).groupby(
    by='cloud',
    as_index=False
).sum().sort_values('count', ascending=False)

fig, ax = plt.subplots(figsize=(12, 6))

sns.barplot(ax=ax, x="cloud", y="count", data=results_df)
fmt_barplot(ax, results_df["count"], results_df['count'].sum())

ax.set_xlabel("Cloud Platform")
ax.set_ylabel("Count")

plt.title(f"Cloud Platform Distribution of All Peers (Total {fmt_thousands(results_df['count'].sum())})")

plt.show()

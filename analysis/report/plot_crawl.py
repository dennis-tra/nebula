import pandas as pd
import seaborn as sns
from matplotlib import pyplot as plt
import matplotlib.dates as md

from lib_db import DBClient, calendar_week
from lib_fmt import thousands_ticker_formatter

sns.set_theme()

client = DBClient()
results = client.query(
    """
    SELECT extract(epoch from started_at), crawled_peers, dialable_peers, undialable_peers
    FROM crawls c
    WHERE c.created_at > date_trunc('week', NOW() - '1 week'::interval)
      AND c.created_at < date_trunc('week', NOW())
    """
)

results_df = pd.DataFrame(results, columns=["started_at", "crawled_peers", "dialable_peers", "undialable_peers"])
results_df['started_at'] = pd.to_datetime(results_df['started_at'], unit='s')
results_df["percentage_dialable"] = 100 * results_df["dialable_peers"] / results_df["crawled_peers"]

fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 10))

sns.lineplot(ax=ax1, x=results_df["started_at"], y=results_df["crawled_peers"])
sns.lineplot(ax=ax1, x=results_df["started_at"], y=results_df["dialable_peers"])
sns.lineplot(ax=ax1, x=results_df["started_at"], y=results_df["undialable_peers"])

ax1.legend(loc='lower right', labels=["Total", "Dialable", "Undialable"])

ax1.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
ax1.set_ylim(0)

ax1.set_xlabel("Time (CEST)")
ax1.set_ylabel("Count")

ax1.get_yaxis().set_major_formatter(thousands_ticker_formatter)

sns.lineplot(ax=ax2, x=results_df["started_at"], y=results_df["percentage_dialable"])
ax2.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
ax2.set_ylim(0, 100)
ax2.set_xlabel("Time (CEST)")
ax2.set_ylabel("Dialable Peers in %")

plt.tight_layout()
plt.savefig(f"./plots-{calendar_week}/crawl-overview.png")
# plt.show()

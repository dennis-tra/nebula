import pandas as pd
import seaborn as sns
from matplotlib import pyplot as plt
import matplotlib.dates as md

from lib_db import DBClient
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

fig, ax = plt.subplots(figsize=(12, 6))

sns.lineplot(ax=ax, x=results_df["started_at"], y=results_df["crawled_peers"])
sns.lineplot(ax=ax, x=results_df["started_at"], y=results_df["dialable_peers"])
sns.lineplot(ax=ax, x=results_df["started_at"], y=results_df["undialable_peers"])

ax.legend(loc='lower right', labels=["Total", "Dialable", "Undialable"])

ax.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
ax.set_ylim(0)

ax.set_xlabel("Time")
ax.set_ylabel("Count")

ax.get_yaxis().set_major_formatter(thousands_ticker_formatter)

plt.title("CESTf")
plt.tight_layout()
plt.show()

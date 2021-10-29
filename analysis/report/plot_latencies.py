import pandas as pd
import numpy as np
import seaborn as sns
from matplotlib import pyplot as plt

from lib_fmt import thousands_ticker_formatter
from lib_db import DBClient, calendar_week

sns.set_theme()

client = DBClient()
crawl_results = client.query(
    """
    SELECT
        EXTRACT('epoch' FROM v.connect_duration), 
        EXTRACT('epoch' FROM v.crawl_duration)
    FROM visits v
    WHERE v.created_at > date_trunc('week', NOW() - '1 week'::interval)
      AND v.created_at < date_trunc('week', NOW())
      AND v.type = 'crawl'
      AND v.error IS NULL
    """
)
dial_results = client.query(
    """
    SELECT
        EXTRACT('epoch' FROM v.dial_duration)
    FROM visits v
    WHERE v.created_at > date_trunc('week', NOW() - '1 week'::interval)
      AND v.created_at < date_trunc('week', NOW())
      AND v.type = 'dial'
      AND v.error IS NULL
    """
)

fig, (ax1, ax2, ax3) = plt.subplots(1, 3, figsize=(12,4))

crawl_results_df = pd.DataFrame(crawl_results, columns=["connect_duration", "crawl_duration"])
dial_results_df = pd.DataFrame(dial_results, columns=["dial_duration"])

sns.histplot(ax=ax1, data=dial_results_df, x="dial_duration", bins=np.arange(0, 10, 0.1))
sns.histplot(ax=ax2, data=crawl_results_df, x="connect_duration", bins=np.arange(0, 40, 0.5))
sns.histplot(ax=ax3, data=crawl_results_df, x="crawl_duration", bins=np.arange(0, 40, 0.5))

ax1.get_yaxis().set_major_formatter(thousands_ticker_formatter)
ax2.get_yaxis().set_major_formatter(thousands_ticker_formatter)
ax3.get_yaxis().set_major_formatter(thousands_ticker_formatter)

ax1.set_xlabel("Dial Duration in s")
ax2.set_xlabel("Connect Duration in s")
ax3.set_xlabel("Connect plus Crawl Duration in s")

plt.tight_layout()

plt.savefig(f"./plots-{calendar_week}/latencies.png")
plt.show()

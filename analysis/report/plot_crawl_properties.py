import pandas as pd
import seaborn as sns
from matplotlib import pyplot as plt, ticker

from lib_agent import agent_name, go_ipfs_v08_version, go_ipfs_version, known_agents
from lib_db import DBClient
from lib_fmt import thousands_ticker_formatter, fmt_thousands

sns.set_theme()

client = DBClient()
results = client.query(
    """
    SELECT cp.crawl_id, EXTRACT('epoch' FROM c.started_at) started_at, av.agent_version, cp.count
    FROM crawl_properties cp 
        INNER JOIN agent_versions av ON cp.agent_version_id = av.id
        INNER JOIN crawls c ON cp.crawl_id = c.id
    WHERE cp.created_at > date_trunc('week', NOW() - '1 week'::interval)
      AND cp.created_at < date_trunc('week', NOW())
      AND cp.count > 10
    """
)

results_df = pd.DataFrame(results, columns=['crawl_id', 'started_at', 'agent_version', 'count']).assign(
    agent_name=lambda df: df.agent_version.apply(agent_name),
    go_ipfs_version=lambda df: df.agent_version.apply(go_ipfs_version),
    go_ipfs_v08_version=lambda df: df.agent_version.apply(go_ipfs_v08_version)
)
results_df['started_at'] = pd.to_datetime(results_df['started_at'], unit='s')

fig, axs = plt.subplots(len(known_agents), 1, figsize=(12, len(known_agents)*3), sharex=True)

for idx, agent in enumerate(known_agents):
    ax = axs[idx]

    grouped_df = results_df \
        .groupby(by=['crawl_id', 'started_at', 'agent_name'], as_index=False) \
        .sum() \
        .sort_values('count', ascending=False)

    values = grouped_df[grouped_df['agent_name'] == agent]['count']
    sns.lineplot(ax=ax, x=grouped_df['started_at'], y=values)
    ax.set_ylim(0)
    ax.set_xlabel("Time (CEST)")
    ax.set_ylabel("Count")
    ax.title.set_text(f"Dialable peers with agent '{agent}'")

    if values.max() < 2000:
        pass
    elif values.max() < 4500:
        ax.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: "%.1fk" % (x / 1000)))
    else:
        ax.get_yaxis().set_major_formatter(thousands_ticker_formatter)

plt.tight_layout()
plt.show()

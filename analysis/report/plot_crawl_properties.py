import pandas as pd
import seaborn as sns
import lib_plot
from matplotlib import pyplot as plt, ticker

from lib_agent import agent_name, go_ipfs_v08_version, go_ipfs_version, known_agents
from lib_db import DBClient
from lib_fmt import thousands_ticker_formatter


def main(db_client: DBClient):
    sns.set_theme()

    results = db_client.get_crawl_properties()

    results_df = pd.DataFrame(results, columns=['crawl_id', 'started_at', 'agent_version', 'count']).assign(
        agent_name=lambda df: df.agent_version.apply(agent_name),
        go_ipfs_version=lambda df: df.agent_version.apply(go_ipfs_version),
        go_ipfs_v08_version=lambda df: df.agent_version.apply(go_ipfs_v08_version)
    )
    results_df['started_at'] = pd.to_datetime(results_df['started_at'], unit='s')

    grouped_df = results_df \
        .groupby(by=['crawl_id', 'started_at', 'agent_name'], as_index=False) \
        .sum() \
        .sort_values('count', ascending=False)

    fig, axs = plt.subplots(1, 2, figsize=(15, 5), sharex=True)

    labels = []
    for idx, agent in enumerate(known_agents):
        if idx == 0:
            ax = axs[0]
            ax.set_ylim(0, grouped_df[grouped_df['agent_name'] == agent]['count'].max() * 1.1)
        else:
            ax = axs[1]
            labels += [agent]

        values = grouped_df[grouped_df['agent_name'] == agent]['count']
        sns.lineplot(ax=ax, x=grouped_df['started_at'], y=values)
        ax.set_xlabel("Time (CEST)")
        ax.set_ylabel("Count")

        if values.max() < 2000:
            pass
        elif values.max() < 4500:
            ax.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: "%.1fk" % (x / 1000)))
        else:
            ax.get_yaxis().set_major_formatter(thousands_ticker_formatter)

    axs[0].legend(loc="lower left", labels=[known_agents[0]])
    axs[1].legend(loc="upper left", labels=labels)

    fig.suptitle(f"Dialable Peers by Agent")

    plt.tight_layout()
    lib_plot.savefig("crawl-properties")
    plt.show()


if __name__ == '__main__':
    db_client = DBClient()
    main(db_client)

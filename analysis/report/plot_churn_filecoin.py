import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
from matplotlib import ticker

import lib_plot
from lib_db_filecoin import DBClientFilecoin
from lib_agent import agent_name, go_ipfs_version_mappings, go_ipfs_version
from lib_db import DBClient
from lib_fmt import fmt_thousands


def main(client: DBClient):
    sns.set_theme()

    results = client.get_node_uptime()

    df = pd.DataFrame(results, columns=['uptime_in_s', 'agent_version'])
    df = df.assign(
        uptime_in_h=df.uptime_in_s.apply(lambda x: x / 3600),
    )

    def configure_axis(ax):
        ax.set_xlim(0, 48)
        ax.set_xticks(np.arange(0, 49, step=2))
        ax.set_xlabel("Uptime in Hours")
        ax.set_ylabel("Online Peers in %")
        ax.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: "%d" % int(x * 100)))

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(15, 5))  # rows, cols

    sns.ecdfplot(ax=ax1, x="uptime_in_h", data=df)
    ax1.legend(loc='lower right', labels=[f"all ({fmt_thousands(len(df))})"])
    configure_axis(ax1)

    agent_labels = []

    results = client.get_agent_versions_distribution()
    top_agent_versions = [result[0] for result in results[:4]]

    for agent in top_agent_versions:
        data = df[df['agent_version'].str.contains(agent)]
        agent_labels += [f"{agent} ({fmt_thousands(len(data))})"]
        sns.ecdfplot(ax=ax2, x="uptime_in_h", data=data)

    ax2.legend(loc='lower right', labels=agent_labels)
    configure_axis(ax2)

    fig.suptitle(f"Node Churn Rate (Total Sessions {fmt_thousands(len(df))})")

    fig.tight_layout()
    lib_plot.savefig("crawl-churn")
    fig.show()


if __name__ == '__main__':
    client = DBClientFilecoin()
    main(client)

import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
from matplotlib import ticker

from lib import lib_plot
from lib.lib_agent import agent_name, known_agents, go_ipfs_version_mappings, go_ipfs_version
from lib.lib_db import DBClient
from lib.lib_fmt import fmt_thousands


def main(client: DBClient):
    sns.set_theme()

    results = client.get_node_uptime()

    df = pd.DataFrame(results, columns=['uptime_in_s', 'agent_version'])
    df = df.assign(
        uptime_in_h=df.uptime_in_s.apply(lambda x: x / 3600),
        agent_name=df.agent_version.apply(agent_name),
        go_ipfs_version=df.agent_version.apply(go_ipfs_version),
    )

    def configure_axis(ax):
        ax.set_xlim(0, 24)
        ax.set_xticks(np.arange(0, 25, step=2))
        ax.set_xlabel("Uptime in Hours")
        ax.set_ylabel("Online Peers in %")
        ax.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: "%d" % int(x * 100)))

    fig, (ax1, ax2, ax3) = plt.subplots(1, 3, figsize=(15, 5))  # rows, cols

    sns.ecdfplot(ax=ax1, x="uptime_in_h", data=df)
    ax1.legend(loc='lower right', labels=[f"all ({fmt_thousands(len(df))})"])
    configure_axis(ax1)

    agent_labels = []
    all_agents = known_agents + ['others']
    for agent in all_agents:
        data = df[df['agent_name'] == agent]
        agent_labels += [f"{agent} ({fmt_thousands(len(data))})"]
        sns.ecdfplot(ax=ax2, x="uptime_in_h", data=data)

    ax2.legend(loc='lower right', labels=agent_labels)
    configure_axis(ax2)

    go_ipfs_version_labels = []
    for go_ipfs_ver in go_ipfs_version_mappings:
        version = go_ipfs_ver[1]
        data = df[df['go_ipfs_version'] == version]
        go_ipfs_version_labels += [f"{version} ({fmt_thousands(len(data))})"]
        sns.ecdfplot(ax=ax3, x="uptime_in_h", data=data)

    ax3.legend(loc='lower right', labels=go_ipfs_version_labels)
    configure_axis(ax3)

    fig.suptitle(f"Node Churn Rate (Total Sessions {fmt_thousands(len(df))})")

    fig.tight_layout()
    lib_plot.savefig("crawl-churn")
    fig.show()


if __name__ == '__main__':
    client = DBClient()
    main(client)

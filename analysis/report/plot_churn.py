import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
from matplotlib import ticker
from lib_agent import agent_name, known_agents, go_ipfs_version_mappings, go_ipfs_version
from lib_db import DBClient
from lib_plot import fmt_thousands

sns.set_theme()

client = DBClient()
results = client.get_node_uptime()

results_df = pd.DataFrame(results, columns=['uptime_in_s', 'agent_version']).assign(
    uptime_in_h=lambda df: df.uptime_in_s.apply(lambda x: x / 3600),
    agent_name=lambda df: df.agent_version.apply(agent_name),
    go_ipfs_version=lambda df: df.agent_version.apply(go_ipfs_version),
)


def customize_axis(ax):
    ax.set_xlim(0, 24)
    ax.set_xticks(np.arange(0, 25, step=2))
    ax.set_xlabel("Uptime in Hours")
    ax.set_ylabel("Online Peers in %")
    ax.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: "%d" % int(x * 100)))


fig, (ax1, ax2, ax3) = plt.subplots(1, 3)  # rows, cols
fig.set_size_inches(15, 5)

sns.ecdfplot(ax=ax1, x="uptime_in_h", data=results_df)
ax1.legend(loc='lower right', labels=[f"all ({fmt_thousands(len(results_df))})"])
customize_axis(ax1)

agent_labels = []
all_agents = known_agents + ['others']
for agent in all_agents:
    data = results_df[results_df['agent_name'] == agent]
    agent_labels += [f"{agent} ({fmt_thousands(len(data))})"]
    sns.ecdfplot(ax=ax2, x="uptime_in_h", data=data)

ax2.legend(loc='lower right', labels=agent_labels)
customize_axis(ax2)

go_ipfs_version_labels = []
for go_ipfs_version in go_ipfs_version_mappings:
    version = go_ipfs_version[1]
    data = results_df[results_df['go_ipfs_version'] == version]
    go_ipfs_version_labels += [f"{version} ({fmt_thousands(len(data))})"]
    sns.ecdfplot(ax=ax3, x="uptime_in_h", data=data)

ax3.legend(loc='lower right', labels=go_ipfs_version_labels)
customize_axis(ax3)

fig.suptitle(f"Node Churn Rate (Total Sessions {fmt_thousands(len(results_df))})")
fig.tight_layout()
fig.show()

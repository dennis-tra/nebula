import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd
from matplotlib import ticker

from lib.lib_agent import agent_name
from lib.lib_fmt import fmt_thousands, fmt_percentage


def plot_cloud_agents(df: pd.DataFrame, clouds: pd.DataFrame) -> plt.Figure:
    df = df.assign(
        agent_name=lambda data_frame: data_frame.agent_version.apply(agent_name),
    )

    unique = df["agent_name"].unique()
    fig, axs = plt.subplots((len(unique) + 1) // 2, 2, figsize=[15, 13], dpi=150)
    for idx, agent in enumerate(sorted(unique)):
        ax = fig.axes[idx]

        data = clouds[clouds["peer_id"].isin(df[df['agent_name'] == agent]["peer_id"])]
        data = data.groupby(by="datacenter", as_index=False).count().sort_values('peer_id',
                                                                                 ascending=False).reset_index(drop=True)
        data = data.rename(columns={'peer_id': 'count'})

        result = data.nlargest(15, columns="count")
        other_count = data.loc[~data["datacenter"].isin(result["datacenter"]), "count"].sum()
        if other_count > 0:
            result.loc[len(result)] = ["Other Datacenters", other_count]

        sns.barplot(ax=ax, x="count", y="datacenter", data=result)
        ax.bar_label(ax.containers[0], list(map(fmt_percentage(result["count"].sum()), result["count"])))
        ax.get_xaxis().set_major_formatter(ticker.StrMethodFormatter('{x:,.0f}'))
        ax.set_xlabel("Count")
        ax.set_ylabel("")
        ax.title.set_text(f"{agent} (Total {fmt_thousands(data['count'].sum())})")

    fig.suptitle(f"Datacenters by Agent Version")
    fig.set_tight_layout(True)

    if len(unique) < len(fig.axes):
        for ax in fig.axes[len(unique):]:
            fig.delaxes(ax)

    return fig

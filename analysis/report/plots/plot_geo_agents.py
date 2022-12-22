import seaborn as sns
import pandas as pd
import matplotlib.pyplot as plt

from lib.lib_agent import agent_name, kubo_version
from lib.lib_fmt import fmt_thousands, fmt_barplot


def plot_geo_agents(df: pd.DataFrame, countries: pd.DataFrame) -> plt.Figure:
    df = df.assign(
        agent_name=lambda data_frame: data_frame.agent_version.apply(agent_name),
        kubo_version=lambda data_frame: data_frame.agent_version.apply(kubo_version),
    )

    fig, axs = plt.subplots(2, 3, figsize=(15, 9))
    for idx, agent in enumerate(sorted(df["agent_name"].unique())):
        ax = fig.axes[idx]

        data = countries[countries["peer_id"].isin(df[df['agent_name'] == agent]["peer_id"])]
        data = data.groupby(by="country", as_index=False).count().sort_values('peer_id', ascending=False)
        data = data.rename(columns={'peer_id': 'count'})

        result = data.nlargest(8, columns="count")
        result.loc[len(result)] = ['Rest', data.loc[~data["country"].isin(result["country"]), "count"].sum()]

        sns.barplot(ax=ax, x="country", y="count", data=result)
        fmt_barplot(ax, result["count"], result["count"].sum())
        ax.set_xlabel("")
        ax.title.set_text(f"{agent} (Total {fmt_thousands(data['count'].sum())})")

    fig.suptitle(f"Country Distributions of all Resolved Peer IDs by Agent Version")
    fig.set_tight_layout(True)

    return fig

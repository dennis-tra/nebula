import pandas as pd
import seaborn as sns
from matplotlib import pyplot as plt, ticker

from lib.lib_agent import agent_name, kubo_version


def plot_crawl_properties(df: pd.DataFrame) -> plt.Figure:
    df = df.assign(
        agent_name=lambda data_frame: data_frame.agent_version.apply(agent_name),
        kubo_version=lambda data_frame: data_frame.agent_version.apply(kubo_version),
    )

    group = df \
        .groupby(by=['crawl_id', 'started_at', 'agent_name'], as_index=False) \
        .sum(numeric_only=True) \
        .sort_values('count', ascending=False)

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=[15, 5], dpi=300)
    sns.lineplot(group, ax=ax1, x="started_at", y="count", hue="agent_name")

    df = df.dropna().assign(
        minor=lambda data_frame: data_frame.kubo_version.apply(lambda row: int(row.split(".")[1])),
    )
    group = df \
        .groupby(by=['crawl_id', 'started_at', 'minor'], as_index=False) \
        .sum(numeric_only=True) \
        .sort_values('count', ascending=False)

    for minor in reversed(sorted(group["minor"].unique())):
        data = group[group["minor"] == minor].sort_values('started_at', ascending=False)
        ax2.plot(data["started_at"], data["count"], label=f"0.{minor}.x")

    ax2.legend(title="kubo")
    fig.set_tight_layout(True)

    return fig

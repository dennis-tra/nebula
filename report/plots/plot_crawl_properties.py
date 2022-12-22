import pandas as pd
import matplotlib.dates as md
from matplotlib import pyplot as plt, ticker

from lib.lib_agent import agent_name, kubo_version
from lib.lib_fmt import thousands_ticker_formatter


def plot_crawl_properties(df: pd.DataFrame) -> plt.Figure:
    df = df.assign(
        agent_name=lambda data_frame: data_frame.agent_version.apply(agent_name),
        kubo_version=lambda data_frame: data_frame.agent_version.apply(kubo_version),
    )

    group = df \
        .groupby(by=['crawl_id', 'started_at', 'agent_name'], as_index=False) \
        .sum(numeric_only=True) \
        .sort_values(['started_at', 'count'], ascending=False)

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=[15, 5], dpi=300)

    for name in sorted(group["agent_name"].unique()):
        data = group[group["agent_name"] == name]
        ax1.plot(data["started_at"], data["count"], label=name)

    ax1.set_ylim(0)
    ax1.set_xlabel("Date")
    ax1.set_ylabel("Count")
    ax1.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
    for tick in ax1.get_xticklabels():
        tick.set_rotation(20)
        tick.set_ha('right')
    ax1.legend(title="Agent Name")
    ax1.set_title("General Agent Versions")
    ax1.get_yaxis().set_major_formatter(thousands_ticker_formatter)

    df = df.dropna().assign(
        minor=lambda data_frame: data_frame.kubo_version.apply(lambda row: int(row.split(".")[1])),
    )
    group = df \
        .groupby(by=['crawl_id', 'started_at', 'minor'], as_index=False) \
        .sum(numeric_only=True) \
        .sort_values('count', ascending=False)

    # Find 10 most widely used agent versions
    filter_group = group \
        .groupby(by="minor", as_index=False) \
        .mean(numeric_only=True) \
        .sort_values('count', ascending=False)
    filter_group = filter_group.head(10)

    for minor in reversed(sorted(group["minor"].unique())):
        if minor not in set(filter_group["minor"]):
            continue
        data = group[group["minor"] == minor].sort_values('started_at', ascending=False)
        ax2.plot(data["started_at"], data["count"], label=f"0.{minor}.x")

    ax2.set_ylim(0)
    ax2.set_xlabel("Date")
    ax2.set_ylabel("Count")
    ax2.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
    for tick in ax2.get_xticklabels():
        tick.set_rotation(20)
        tick.set_ha('right')
    ax2.legend()
    ax2.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: "%.1fk" % (x / 1000)))

    ax2.set_title("Kubo Versions")
    ax2.legend(handlelength=1.0, ncols=5)
    fig.set_tight_layout(True)

    return fig

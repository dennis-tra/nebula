import pandas as pd
import seaborn as sns
import matplotlib.dates as md
from matplotlib import pyplot as plt, ticker

from lib.lib_fmt import thousands_ticker_formatter


def plot_crawl_protocols(df: pd.DataFrame) -> plt.Figure:
    top_protocols = df \
        .groupby("protocol") \
        .sum(numeric_only=True) \
        .sort_values("count", ascending=False) \
        .nlargest(20, columns="count") \
        .reset_index()["protocol"]

    fig, ax = plt.subplots(figsize=[15, 8], dpi=150)

    df = df.fillna(0)

    colors = sns.color_palette()
    for idx, protocol in enumerate(top_protocols.to_numpy()):
        linestyle = '-'
        if idx >= 10:
            linestyle = "--"

        color = colors[idx % 10]
        filtered = df[df["protocol"] == protocol]
        ax.plot(filtered["started_at"], filtered["count"], label=protocol, linestyle=linestyle, color=color)

    ax.set_xlabel("Date")
    ax.set_ylabel("Count")
    ax.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
    for tick in ax.get_xticklabels():
        tick.set_rotation(20)
        tick.set_ha('right')
    ax.legend(ncols=5, loc="upper center", bbox_to_anchor=(0.5, 1.2))
    ax.get_yaxis().set_major_formatter(ticker.StrMethodFormatter('{x:,.0f}'))

    fig.set_tight_layout(True)

    return fig

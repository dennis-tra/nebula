import pandas as pd
import seaborn as sns
from matplotlib import pyplot as plt, ticker
import matplotlib.dates as md


def plot_crawl_unresponsive(df: pd.DataFrame) -> plt.Figure:

    fig, ax = plt.subplots(figsize=[12, 7], dpi=150)

    ax.plot(df["started_at"], df["total"], label="Total Count")
    ax.plot(df["started_at"], df["unresponsive"], label="Unresponsive Count")
    ax.plot(df["started_at"], df["exposed"], label="Exposed Count")
    ax.set_xlabel("Date")
    ax.legend(loc="lower left")
    ax.set_ylabel("Count")
    ax.set_ylim(0)
    ax.get_yaxis().set_major_formatter(ticker.StrMethodFormatter('{x:,.0f}'))
    ax.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d %H:%M'))

    for tick in ax.get_xticklabels():
        tick.set_rotation(10)
        tick.set_ha('right')

    twinax = ax.twinx()
    twinax.plot(df["started_at"], 100 * df["exposed"] / df["total"], color=sns.color_palette()[3], label="Share Exposed")
    twinax.plot(df["started_at"], 100 * df["unresponsive"] / df["total"], color=sns.color_palette()[4], label="Share Unresponsive")
    twinax.set_ylim(0, 100)
    twinax.set_ylabel("Share of Total in %")
    twinax.legend(loc="lower right")
    twinax.grid(False)

    fig.set_tight_layout(True)

    return fig

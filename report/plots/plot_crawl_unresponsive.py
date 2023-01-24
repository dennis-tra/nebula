import pandas as pd
from matplotlib import pyplot as plt, ticker


def plot_crawl_unresponsive(df: pd.DataFrame) -> plt.Figure:

    fig, ax = plt.subplots(figsize=[12, 7], dpi=150)

    ax.plot(df["started_at"], df["unresponsive"], label="Unresponsive")
    ax.plot(df["started_at"], df["total"], label="Total")
    ax.set_xlabel("Date")
    ax.legend(loc="lower left")
    ax.set_ylabel("Count")
    ax.set_ylim(0)
    ax.get_yaxis().set_major_formatter(ticker.StrMethodFormatter('{x:,.0f}'))

    for tick in ax.get_xticklabels():
        tick.set_rotation(10)
        tick.set_ha('right')

    twinax = ax.twinx()
    twinax.plot(df["started_at"], 100 * df["unresponsive"] / df["total"], color="k", label="Share")
    twinax.set_ylim(0, 100)
    twinax.set_ylabel("Share in %")
    twinax.legend(loc="lower right")
    twinax.grid(False)

    fig.set_tight_layout(True)

    return fig

import pandas as pd
from matplotlib import pyplot as plt
import matplotlib.dates as md
from lib.lib_fmt import thousands_ticker_formatter


def plot_crawl_overview(df: pd.DataFrame) -> plt.Figure:

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=[15, 5], dpi=300)

    ax1.plot(df["started_at"], df["crawled_peers"], label="Total")
    ax1.plot(df["started_at"], df["dialable_peers"], label="Dialable")
    ax1.plot(df["started_at"], df["undialable_peers"], label="Undialable")

    ax1.legend(loc='lower right', labels=["Total", "Dialable", "Undialable"])

    for tick in ax1.get_xticklabels():
        tick.set_rotation(20)
        tick.set_ha('right')

    ax1.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
    ax1.set_ylim(0)

    ax1.set_xlabel("Date")
    ax1.set_ylabel("Count")

    ax1.get_yaxis().set_major_formatter(thousands_ticker_formatter)

    ax2.plot(df["started_at"], df["percentage_dialable"])
    ax2.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
    ax2.set_ylim(0, 100)
    ax2.set_xlabel("Time")
    ax2.set_ylabel("Dialable Peers in %")

    for tick in ax2.get_xticklabels():
        tick.set_rotation(20)
        tick.set_ha('right')

    fig.set_tight_layout(True)

    return fig

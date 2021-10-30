import pandas as pd
import seaborn as sns
from matplotlib import pyplot as plt
import matplotlib.dates as md
import lib_plot
from lib_db import DBClient
from lib_fmt import thousands_ticker_formatter


def main():
    sns.set_theme()

    client = DBClient()
    results = client.get_crawls()

    results_df = pd.DataFrame(results, columns=["started_at", "crawled_peers", "dialable_peers", "undialable_peers"])
    results_df['started_at'] = pd.to_datetime(results_df['started_at'], unit='s')
    results_df["percentage_dialable"] = 100 * results_df["dialable_peers"] / results_df["crawled_peers"]

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(15, 5))

    sns.lineplot(ax=ax1, x=results_df["started_at"], y=results_df["crawled_peers"])
    sns.lineplot(ax=ax1, x=results_df["started_at"], y=results_df["dialable_peers"])
    sns.lineplot(ax=ax1, x=results_df["started_at"], y=results_df["undialable_peers"])

    ax1.legend(loc='lower right', labels=["Total", "Dialable", "Undialable"])

    ax1.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
    ax1.set_ylim(0)

    ax1.set_xlabel("Time (CEST)")
    ax1.set_ylabel("Count")

    ax1.get_yaxis().set_major_formatter(thousands_ticker_formatter)

    sns.lineplot(ax=ax2, x=results_df["started_at"], y=results_df["percentage_dialable"])
    ax2.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
    ax2.set_ylim(0, 100)
    ax2.set_xlabel("Time (CEST)")
    ax2.set_ylabel("Dialable Peers in %")

    plt.tight_layout()
    lib_plot.savefig("crawl-overview.png")
    plt.show()


if __name__ == '__main__':
    main()

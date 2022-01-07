import pandas as pd
import numpy as np
import seaborn as sns
from matplotlib import pyplot as plt

import lib_plot
from lib_fmt import thousands_ticker_formatter
from lib_db import DBClient


def main(db_client: DBClient):
    sns.set_theme()

    crawl_results = db_client.get_crawl_visit_durations()
    dial_results = db_client.get_dial_visit_durations()

    fig, (ax1, ax2, ax3) = plt.subplots(1, 3, figsize=(12, 4))

    crawl_results_df = pd.DataFrame(crawl_results, columns=["connect_duration", "crawl_duration"])
    dial_results_df = pd.DataFrame(dial_results, columns=["dial_duration"])

    sns.histplot(ax=ax1, data=dial_results_df, x="dial_duration", bins=np.arange(0, 10, 0.1))
    sns.histplot(ax=ax2, data=crawl_results_df, x="connect_duration", bins=np.arange(0, 40, 0.5))
    sns.histplot(ax=ax3, data=crawl_results_df, x="crawl_duration", bins=np.arange(0, 40, 0.5))

    ax1.get_yaxis().set_major_formatter(thousands_ticker_formatter)
    ax2.get_yaxis().set_major_formatter(thousands_ticker_formatter)
    ax3.get_yaxis().set_major_formatter(thousands_ticker_formatter)

    ax1.set_xlabel("Dial Duration in s")
    ax2.set_xlabel("Connect Duration in s")
    ax3.set_xlabel("Connect plus Crawl Duration in s")

    plt.tight_layout()
    lib_plot.savefig("latencies")
    plt.show()


if __name__ == '__main__':
    db_client = DBClient()
    main(db_client)

import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

from lib_cloud import Cloud
from lib_fmt import fmt_barplot, fmt_thousands

sns.set_theme()


def plot_cloud(data, classification):
    cloud_client = Cloud()

    results_df = pd.DataFrame(data, columns=["ip_address"]).assign(
        cloud=lambda df: df.ip_address.apply(lambda ip: cloud_client.cloud_for(ip)),
        count=lambda df: df.ip_address.apply(lambda ip: 1),
    ).groupby(
        by='cloud',
        as_index=False
    ).sum().sort_values('count', ascending=False)

    fig, ax = plt.subplots(figsize=(12, 6))

    sns.barplot(ax=ax, x="cloud", y="count", data=results_df)
    fmt_barplot(ax, results_df["count"], results_df['count'].sum())

    ax.set_xlabel("Cloud Platform")
    ax.set_ylabel("Count")

    plt.title(f"Cloud Platform Distribution of {classification} Peers (Total {fmt_thousands(results_df['count'].sum())})")

    plt.show()

import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd
from matplotlib import ticker

from lib.lib_fmt import fmt_thousands, fmt_percentage


def plot_cloud_overall(df: pd.DataFrame) -> plt.Figure:
    result = df.nlargest(20, columns="count")
    result.loc[len(result)] = [df.loc[~df["datacenter"].isin(result["datacenter"]), "count"].sum(), "Other Datacenters"]

    fig, ax = plt.subplots(figsize=[15, 5], dpi=150)

    sns.barplot(ax=ax, x="count", y="datacenter", data=result)
    ax.bar_label(ax.containers[0], list(map(fmt_percentage(result["count"].sum()), result["count"])))

    ax.get_xaxis().set_major_formatter(ticker.StrMethodFormatter('{x:,.0f}'))

    ax.set_xlabel("Count")
    ax.set_ylabel("")

    fig.suptitle(f"Datacenter Distribution of All Peers (Total {fmt_thousands(df['count'].sum())})")

    fig.set_tight_layout(True)

    return fig

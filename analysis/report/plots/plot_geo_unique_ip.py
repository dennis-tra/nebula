import seaborn as sns
import pandas as pd
from matplotlib import pyplot as plt

from lib.lib_fmt import fmt_barplot, thousands_ticker_formatter, fmt_thousands


def plot_geo_unique_ip(df: pd.DataFrame) -> plt.Figure:
    result = df.nlargest(20, columns="count")
    result.loc[len(result)] = ['Rest', df.loc[~df["country"].isin(result["country"]), "count"].sum()]

    fig, ax = plt.subplots(figsize=[15, 5], dpi=300)

    sns.barplot(ax=ax, x="country", y="count", data=result)
    fmt_barplot(ax, result["count"], result["count"].sum())
    ax.set_xlabel("")
    ax.set_ylabel("Count")
    ax.get_yaxis().set_major_formatter(thousands_ticker_formatter)

    plt.title(f"Country Distribution of Unique IP Addresses (Total {fmt_thousands(result['count'].sum())})")

    fig.set_tight_layout(True)

    return fig

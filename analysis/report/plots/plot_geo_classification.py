import seaborn as sns
import matplotlib.pyplot as plt
from matplotlib import ticker

from lib.lib_fmt import fmt_thousands, fmt_barplot


def plot_geo_classification(distributions) -> plt.Figure:
    fig, axs = plt.subplots(2, 3, figsize=(15, 8))

    for idx, node_class in enumerate(distributions):
        ax = axs[idx // 3][idx % 3]

        data = distributions[node_class]
        result = data.nlargest(8, columns="count")
        result.loc[len(result)] = ['Rest', data.loc[~data["country"].isin(result["country"]), "count"].sum()]

        sns.barplot(ax=ax, x="country", y="count", data=result)
        fmt_barplot(ax, result["count"], result["count"].sum())
        ax.set_xlabel("")
        ax.set_ylabel("Count")
        ax.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: "%.1fk" % (x / 1000)))

        ax.title.set_text(f"{node_class.value} (Total {fmt_thousands(data['count'].sum())})")

    fig.suptitle(f"Country Distributions by Node Classification")

    fig.set_tight_layout(True)

    return fig

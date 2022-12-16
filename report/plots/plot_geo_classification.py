import seaborn as sns
import matplotlib.pyplot as plt
from matplotlib import ticker

from lib.lib_fmt import fmt_thousands, fmt_percentage


def plot_geo_classification(distributions) -> plt.Figure:
    fig, axs = plt.subplots(2, 3, figsize=[15, 8], dpi=150)

    for idx, node_class in enumerate(distributions):
        ax = axs[idx // 3][idx % 3]

        data = distributions[node_class]
        result = data.nlargest(8, columns="count")
        other_count = data.loc[~data["country"].isin(result["country"]), "count"].sum()

        if other_count > 0:
            result.loc[len(result)] = ['Rest', other_count]

        if len(result) > 0:
            sns.barplot(ax=ax, x="country", y="count", data=result)
            ax.bar_label(ax.containers[0], list(map(fmt_percentage(result["count"].sum()), result["count"])))

        ax.set_xlabel("")
        ax.set_ylabel("Count")
        ax.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: "%.1fk" % (x / 1000)))

        ax.title.set_text(f"{node_class.value} (Total {fmt_thousands(data['count'].sum())})")

    fig.suptitle(f"Country Distributions by Node Classification")

    fig.set_tight_layout(True)

    if len(distributions) < len(fig.axes):
        for ax in fig.axes[len(distributions):]:
            fig.delaxes(ax)

    return fig

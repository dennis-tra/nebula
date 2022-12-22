import seaborn as sns
import matplotlib.pyplot as plt

from lib.lib_fmt import fmt_thousands, thousands_ticker_formatter, fmt_percentage


def plot_cloud_classification(distributions) -> plt.Figure:
    fig, axs = plt.subplots(3, 2, figsize=[15, 10], dpi=150)

    for idx, node_class in enumerate(distributions):
        ax = fig.axes[idx]
        data = distributions[node_class]

        result = data.nlargest(15, columns="count")
        other_count = data.loc[~data["datacenter"].isin(result["datacenter"]), "count"].sum()
        if other_count > 0:
            result.loc[len(result)] = ["Other Datacenters", other_count]

        sns.barplot(ax=ax, x="count", y="datacenter", data=result)
        ax.bar_label(ax.containers[0], list(map(fmt_percentage(result["count"].sum()), result["count"])))
        ax.get_xaxis().set_major_formatter(thousands_ticker_formatter)
        ax.set_xlabel("Count")
        ax.set_ylabel("")
        ax.title.set_text(f"{node_class.name.lower()} (Total {fmt_thousands(data['count'].sum())})")

    fig.suptitle(f"Datacenter Distributions by Node Classification")

    fig.set_tight_layout(True)

    return fig

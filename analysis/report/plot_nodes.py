from collections import OrderedDict

import seaborn as sns
import matplotlib.pyplot as plt
from lib_db import DBClient, NodeClassification, calendar_week
from lib_fmt import fmt_thousands, thousands_ticker_formatter, fmt_percentage


def main():
    sns.set_theme()

    client = DBClient()

    data = OrderedDict()
    for node_class in NodeClassification:
        get_peer_ids = client.node_classification_funcs[node_class]
        data[node_class.value] = len(get_peer_ids())

    # order dict by count decreasing
    data = OrderedDict(reversed(sorted(data.items(), key=lambda item: item[1])))

    all_peer_ids = client.get_all_peer_ids()

    fig, (ax) = plt.subplots(figsize=(8, 5))

    sns.barplot(ax=ax, x=list(data.keys()), y=list(data.values()))
    ax.get_yaxis().set_major_formatter(thousands_ticker_formatter)
    ax.bar_label(ax.containers[0], list(map(fmt_percentage(len(all_peer_ids)), data.values())))

    ax.title.set_text(f"Node Classification of {fmt_thousands(len(all_peer_ids))} Visited Peers")
    ax.set_ylabel("Count")

    plt.tight_layout()

    plt.savefig(f"./plots-{calendar_week}/nodes.png")
    plt.show()


if __name__ == '__main__':
    main()

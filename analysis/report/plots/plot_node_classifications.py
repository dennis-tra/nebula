import seaborn as sns
import matplotlib.pyplot as plt

from lib import NodeClassification, DBClient
from lib.lib_fmt import fmt_thousands, thousands_ticker_formatter, fmt_percentage
from collections import OrderedDict


def data_node_classifications(db_client: DBClient):
    data = OrderedDict()
    for node_class in NodeClassification:
        get_peer_ids = db_client.node_classification_funcs[node_class]
        data[node_class.value] = len(get_peer_ids())

    return db_client.get_all_peer_ids(), OrderedDict(reversed(sorted(data.items(), key=lambda item: item[1])))


def plot_node_classifications(all_peer_ids: list[int], data: OrderedDict) -> plt.Figure:
    fig, ax = plt.subplots(figsize=[10, 5], dpi=300)
    sns.barplot(ax=ax, x=list(data.keys()), y=list(data.values()))
    ax.get_yaxis().set_major_formatter(thousands_ticker_formatter)
    ax.bar_label(ax.containers[0], list(map(fmt_percentage(len(all_peer_ids)), data.values())))
    ax.title.set_text(f"Node Classification of {fmt_thousands(len(all_peer_ids))} Visited Peers")
    ax.set_ylabel("Count")

    fig.set_tight_layout(True)

    return fig

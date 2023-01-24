from typing import Dict, Optional

import pandas as pd
import seaborn as sns
import matplotlib.dates as md
from matplotlib import pyplot as plt

from lib import NodeClassification
from lib.lib_fmt import thousands_ticker_formatter


def plot_crawl_classifications(data: Dict[NodeClassification, pd.DataFrame]) -> plt.Figure:

    combined: Optional[pd.DataFrame] = None
    for classification in data.keys():
        if combined is None:
            combined = data[classification]
            combined[f"count_{classification.name.lower()}"] = combined["count"]
            continue

        combined = combined.merge(data[classification], how='outer', on="started_at", suffixes=("", f"_{classification.name.lower()}"))
    combined = combined.fillna(0)

    palette = sns.color_palette()
    color_map = {
        NodeClassification.ONLINE: palette[2],
        NodeClassification.OFFLINE: palette[3],
        NodeClassification.ONEOFF: palette[7],
        NodeClassification.DANGLING: palette[1],
        NodeClassification.ENTERED: palette[0],
        NodeClassification.LEFT: palette[4],
    }

    xs = combined["started_at"]
    ys = []
    ls = []
    cs = []
    for classification in data.keys():
        ys += [combined[f"count_{classification.name.lower()}"]]
        ls += [classification.name]
        cs += [color_map[classification]]

    fig, ax = plt.subplots(figsize=[15, 5], dpi=150)

    ax.stackplot(xs, ys, labels=ls, colors=cs, alpha=0.8)

    ax.set_ylim(0)
    ax.set_xlabel("Date")
    ax.set_ylabel("Count")
    ax.legend(loc="lower center", ncols=2)
    ax.xaxis.set_major_formatter(md.DateFormatter('%Y-%m-%d'))
    for tick in ax.get_xticklabels():
        tick.set_rotation(10)
        tick.set_ha('right')
    ax.get_yaxis().set_major_formatter(thousands_ticker_formatter)

    fig.set_tight_layout(True)

    return fig

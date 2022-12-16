import numpy as np
import matplotlib.pyplot as plt
import pandas as pd
from matplotlib import ticker
from lib.lib_agent import kubo_version
from lib.lib_fmt import fmt_thousands


def plot_agents_kubo(df: pd.DataFrame, include_storm=True):
    storm_df = df[df["is_storm"] == True]
    df = df[df["is_storm"] == False]
    df = df.assign(
        kubo_version=lambda data_frame: data_frame.agent_version.apply(kubo_version),
    ).dropna()

    df = df.assign(
        minor=lambda data_frame: data_frame.kubo_version.apply(lambda row: int(row.split(".")[1])),
    )

    df = df \
        .groupby(by=["minor"], as_index=False) \
        .sum(numeric_only=True) \
        .sort_values(["minor", "count"], ascending=False) \
        .reset_index(drop=True)
    storm_df = storm_df.assign(
        kubo_version=lambda data_frame: data_frame.agent_version.apply(kubo_version),
    ).dropna()

    storm_df = storm_df.assign(
        minor=lambda data_frame: data_frame.kubo_version.apply(lambda row: int(row.split(".")[1])),
    )

    storm_df = storm_df \
        .groupby(by=["minor"], as_index=False) \
        .sum(numeric_only=True) \
        .sort_values(["minor", "count"], ascending=False) \
        .reset_index(drop=True)

    if include_storm:
        kubo_versions_total = df['count'].sum() + storm_df['count'].sum()
    else:
        kubo_versions_total = df['count'].sum()

    fig, ax = plt.subplots(figsize=[10, 5], dpi=300)

    p1 = ax.barh(df["minor"], df["count"])
    ax.set_yticks(df["minor"], labels=[f"0.{minor}.x" for minor in df["minor"]])
    ax.set_xlabel("Count")
    ax.set_ylabel("kubo/go-ipfs Agent")

    ax.bar_label(p1, padding=6, labels=["%.1f%%" % (100*val/kubo_versions_total) for val in df["count"]])

    if include_storm:
        zeros = np.zeros(len(df["count"]))
        left = zeros.copy()
        count = zeros.copy()
        minors = []
        for idx, row in storm_df.iterrows():
            minor_index = df["minor"][df["minor"] == row["minor"]].index[0]
            left[minor_index] += df.iloc[minor_index]["count"]
            count[minor_index] += row["count"]
            minors += [row["minor"]]

        p2 = ax.barh(minors, count, label=f"storm", left=left)
        ax.bar_label(p2, padding=6, labels=["%.1f%%" % (100*val/kubo_versions_total) if val > 0 else "" for val in count])

        ax.legend()

    ax.set_title(f"Agent Types (Total Peers {fmt_thousands(kubo_versions_total)})")
    ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x, p: format(int(x), ',')))
    fig.set_tight_layout(True)

    return fig

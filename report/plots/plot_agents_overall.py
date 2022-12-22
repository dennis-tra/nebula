import numpy as np
import matplotlib.pyplot as plt
import pandas as pd
from matplotlib import ticker
from lib.lib_agent import agent_name, kubo_version
from lib.lib_fmt import fmt_thousands


def plot_agents_overall(df: pd.DataFrame) -> plt.Figure:
    df = df.assign(
        agent_name=lambda data_frame: data_frame.agent_version.apply(agent_name),
    )

    agent_names_df = df \
        .groupby(by=["agent_name", "is_storm"], as_index=False) \
        .sum(numeric_only=True) \
        .sort_values('count', ascending=False)
    agent_names_total = agent_names_df["count"].sum()

    fig, ax = plt.subplots(figsize=[10, 5], dpi=300)

    peers_regular = agent_names_df[(agent_names_df["agent_name"] == "storm") | (agent_names_df["is_storm"] == False)].reset_index(drop=True)
    peers_storm = agent_names_df[(agent_names_df["agent_name"] != "storm") & (agent_names_df["is_storm"] == True)].reset_index(drop=True)

    bar = ax.bar(peers_regular["agent_name"], peers_regular["count"], label="Regular")
    ax.bar_label(bar, padding=6, labels=["%.1f%%" % (100*val/agent_names_total) for val in peers_regular["count"]])

    # find index of storm nodes
    storm_index = peers_regular["agent_name"][peers_regular["agent_name"] == "storm"].index[0]

    zeros = np.zeros(len(peers_regular["count"]))
    bottom = zeros.copy()
    bottom[storm_index] = peers_regular.iloc[storm_index]["count"]
    for idx, row in peers_storm.iterrows():
        count = zeros.copy()
        count[storm_index] = row["count"]
        bar = ax.bar(peers_regular["agent_name"], count, label=f"{row['agent_name']}", bottom=bottom)
        ax.bar_label(bar, padding=6, labels=["%.1f%%" % (100 * val / agent_names_total) if val > 0 else "" for val in count])
        bottom[storm_index] += row["count"]

    ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x, p: format(int(x), ',')))
    ax.legend(title="Node Type")
    ax.set_xlabel("Agent")
    ax.set_ylabel("Count")
    ax.set_title(f"Agent Types (Total Peers {fmt_thousands(agent_names_total)})")

    fig.set_tight_layout(True)

    return fig

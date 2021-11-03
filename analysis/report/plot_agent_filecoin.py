import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

import lib_plot
from lib_db_filecoin import DBClientFilecoin
from lib_db import DBClient, NodeClassification
from lib_fmt import fmt_barplot, fmt_thousands


def main(client: DBClient):
    sns.set_theme()

    def plot_agent(results, plot_name, threshold):
        results_df = pd.DataFrame(results, columns=['agent_version', 'count']).assign(
            version=lambda df: df.agent_version.apply(lambda av: av[6:]),
        )
        results_total = results_df['count'].sum()

        granular_df = results_df[results_df["count"] > threshold]
        others_df = results_df[results_df["count"] <= threshold]
        others_sum_df = pd.DataFrame([["other", others_df["count"].sum()]], columns=["version", "count"])
        all_df = granular_df.append(others_sum_df)

        # Plotting
        fig, ax = plt.subplots(figsize=(15, 5))  # rows, cols

        sns.barplot(ax=ax, x='version', y='count', data=all_df)
        fmt_barplot(ax, all_df["count"], results_total)
        ax.title.set_text(f"Lotus Agent Versions (Total {fmt_thousands(results_total)})")
        ax.set_xlabel("Agent")
        ax.set_ylabel("Count")

        plt.suptitle(f"Lotus Version Distribution of '{plot_name}' Peers")
        plt.tight_layout()
        lib_plot.savefig(f"agents-{plot_name}")
        plt.show()

    results = client.get_agent_versions_distribution()
    plot_agent(results, "all", 20)

    agent_version_distributions = {}
    thresholds = {
        NodeClassification.ONEOFF: 10,
        NodeClassification.DANGLING: 40,
        NodeClassification.ONLINE: 10,
        NodeClassification.ENTERED: 5,
    }

    for node_class in thresholds.keys():
        peer_ids = client.node_classification_funcs[node_class]()
        agent_version_distributions[node_class] = client.get_agent_versions_for_peer_ids(peer_ids)

    fig, axs = plt.subplots(2, 2, figsize=(15, 9))

    for idx, node_class in enumerate(agent_version_distributions):
        data = agent_version_distributions[node_class]
        ax = axs[idx // 2][idx % 2]

        # calculate the "other" agents
        results_df = pd.DataFrame(data, columns=['agent_version', 'count']).assign(
            version=lambda df: df.agent_version.apply(lambda av: av[6:]),
        )

        granular_df = results_df[results_df["count"] > thresholds[node_class]]
        others_df = results_df[results_df["count"] <= thresholds[node_class]]
        others_sum_df = pd.DataFrame([["other", others_df["count"].sum()]], columns=["version", "count"])
        all_df = granular_df.append(others_sum_df)

        sns.barplot(ax=ax, x="version", y="count", data=all_df)
        fmt_barplot(ax, all_df["count"], all_df["count"].sum())
        ax.set_xlabel("Lotus Agent Version")

        ax.title.set_text(f"{node_class.value} (Total {fmt_thousands(all_df['count'].sum())})")

    plt.suptitle(f"Lotus Agent Version Distribution by Node Classification")
    plt.tight_layout()
    lib_plot.savefig(f"agent-classification-filecoin")
    plt.show()


if __name__ == '__main__':
    client = DBClientFilecoin()
    main(client)

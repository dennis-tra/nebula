import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

from lib_db import calendar_week
from lib_fmt import fmt_thousands, fmt_barplot

sns.set_theme()


def plot_geo(data, classification, threshold, file_name):
    fig, ax = plt.subplots(figsize=(12, 6))

    # calculate the "other" countries
    granular_df = data[data["Count"] > threshold]
    others_df = data[data["Count"] <= threshold]
    others_sum_df = pd.DataFrame([["other", others_df["Count"].sum()]], columns=["Country", "Count"])
    all_df = granular_df.append(others_sum_df)

    sns.barplot(ax=ax, x="Country", y="Count", data=all_df)
    fmt_barplot(ax, all_df["Count"], all_df["Count"].sum())

    plt.title(f"Country Distribution of {classification} Peers (Total {fmt_thousands(data['Count'].sum())})")

    plt.savefig(f"./plots-{calendar_week}/geo-{file_name}.png")
    plt.show()

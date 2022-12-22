import os
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt


def cdf(series: pd.Series) -> pd.DataFrame:
    """ calculates the cumulative distribution function of the given series"""
    return pd.DataFrame.from_dict({
        series.name: np.append(series.sort_values(), series.max()),
        "cdf": np.linspace(0, 1, len(series) + 1)
    })


def savefig(fig: plt.Figure, name: str, calendar_week: int):
    dir_name = f"plots-{calendar_week}"
    if not os.path.isdir(dir_name):
        os.mkdir(dir_name)
    fig.savefig(f"{dir_name}/{name}.png")

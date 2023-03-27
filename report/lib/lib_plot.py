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


def savefig(fig: plt.Figure, name: str, dir_name='plots'):
    fig.savefig(os.path.join(dir_name, f"{name}.png"))

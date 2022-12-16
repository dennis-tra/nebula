import os
import matplotlib.pyplot as plt


def savefig(fig: plt.Figure, name: str, calendar_week: int):
    dir_name = f"plots-{calendar_week}"
    if not os.path.isdir(dir_name):
        os.mkdir(dir_name)
    fig.savefig(f"{dir_name}/{name}.png")

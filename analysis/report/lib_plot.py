import os
import matplotlib.pyplot as plt

from lib_db import calendar_week


def savefig(name: str):
    dir_name = f"plots-{calendar_week}"
    if not os.path.isdir(dir_name):
        os.mkdir(dir_name)
    plt.savefig(f"{dir_name}/{name}.png")

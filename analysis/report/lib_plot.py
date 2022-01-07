from pathlib import Path
import matplotlib.pyplot as plt

from lib_db import calendar_week


def savefig(name: str):
    dir_name = f"reports/plots-{calendar_week}"
    path(dir_name).mkdir(parents=True, exists_ok=True)
    plt.savefig(f"{dir_name}/{name}.png")

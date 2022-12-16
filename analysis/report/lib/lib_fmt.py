from matplotlib import ticker
from datetime import datetime as dt

thousands_ticker_formatter = ticker.FuncFormatter(lambda x, p: "%dk" % int(x / 1000))


def fmt_discovered_entity(data) -> list[str]:
    return [f"`{d[1]}` ({dt.utcfromtimestamp(d[0]).strftime('%Y-%m-%d %H:%M:%S')})" for d in data]


def fmt_thousands(val: int) -> str:
    return format(val, ",")


def fmt_percentage(total: int):
    return lambda val: "%.1f%%" % (100 * val / total)


def fmt_barplot(ax, values, total):
    ax.bar_label(ax.containers[0], list(map(fmt_percentage(total), values)))
    if values.max() < 2000:
        return
    elif values.max() < 4500:
        ax.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: "%.1fk" % (x / 1000)))
    else:
        ax.get_yaxis().set_major_formatter(thousands_ticker_formatter)

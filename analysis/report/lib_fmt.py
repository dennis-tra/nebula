from matplotlib import ticker

thousands_ticker_formatter = ticker.FuncFormatter(lambda x, p: "%dk" % int(x / 1000))


def fmt_thousands(val: int) -> str:
    return format(val, ",")


def fmt_percentage(total: int):
    return lambda val: "%.1f%%" % (100 * val / total)


def fmt_barplot(ax, values, total):
    ax.bar_label(ax.containers[0], list(map(fmt_percentage(total), values)))
    if total > 2000:
        ax.get_yaxis().set_major_formatter(thousands_ticker_formatter)

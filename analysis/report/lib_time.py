import datetime


class Time:
    year = datetime.date.today().isocalendar().year
    calendar_week = (datetime.date.today() - datetime.timedelta(weeks=1)).isocalendar().week
    measurement_start = datetime.datetime.strptime(f"{year}-W{calendar_week}" + '-1', "%Y-W%W-%w").date(),
    measurement_end = datetime.datetime.strptime(f"{year}-W{calendar_week + 1}" + '-1', "%Y-W%W-%w").date(),

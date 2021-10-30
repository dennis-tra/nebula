import datetime


class Time:
    calendar_week = (datetime.date.today() - datetime.timedelta(weeks=1)).isocalendar().week

import pytz


# get_node_uptime gets the session uptime between start and end.
# It takes an sql connection, the start time, the end time as arguments, and
# returns the a list containing the uptime.
def get_node_uptime(conn, start, end):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
            SELECT EXTRACT(EPOCH FROM min_duration)
            FROM sessions
            WHERE created_at < %s AND updated_at > %s
        """,
        [end, start]
    )

    return [i for sub in cur.fetchall() for i in sub]

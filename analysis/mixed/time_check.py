# get_time_range gets the data collection start time and end time.
# It takes an sql connection as the argument, and
# returns the start time and end time (in local timezone) of the data collection.
def get_time_range(conn):
    cur = conn.cursor()
    cur.execute(
        """
        SELECT MIN(updated_at), MAX(updated_at)
        FROM sessions
        """
    )
    records = cur.fetchall()
    return records[0][0].astimezone(), records[0][1].astimezone()

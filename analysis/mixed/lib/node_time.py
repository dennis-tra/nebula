# get_time_range gets the data collection start time and end time.
# It takes an sql connection as the argument, and
# returns the start time and end time (in local timezone) of the data collection.
def get_time_range(conn):
    cur = conn.cursor()
    cur.execute(
        """
        SELECT MIN(created_at), MAX(updated_at)
        FROM sessions
        """
    )
    record = cur.fetchone()
    return record[0].astimezone(), record[1].astimezone()

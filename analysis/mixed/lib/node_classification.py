import pytz


# get_all_nodes gets the id of all nodes between two timestamps.
# It takes an sql connection, the start time, the end time as the arguments, and
# returns the ids of all the nodes.
def get_all_nodes(conn, start, end):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM sessions
        WHERE created_at < %s AND updated_at > %s
        """,
        [end, start]
    )
    return [i for sub in cur.fetchall() for i in sub]


# get_on_nodes gets the id of all on nodes between two timestamps.
# It takes an sql connection, the start time, the end time as the arguments, and
# returns the ids of all the on nodes.
def get_on_nodes(conn, start, end):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM sessions
        WHERE created_at < %s AND updated_at > %s AND first_successful_dial != last_successful_dial AND peer_id NOT IN (
            SELECT peer_id
            FROM sessions
            WHERE updated_at > %s AND updated_At < %s
        )
        """,
        [end, end, start, end]
    )
    return [i for sub in cur.fetchall() for i in sub]


# get_off_nodes gets the id of all off nodes between two timestamps.
# It takes an sql connection, the start time, the end time as the arguments, and
# returns the ids of all the off nodes.
def get_off_nodes(conn, start, end):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM sessions
        WHERE created_at < %s AND updated_at > %s AND first_successful_dial = last_successful_dial AND finished = true AND peer_id NOT IN (
            SELECT peer_id
            FROM sessions
            WHERE updated_at > %s AND updated_At < %s AND first_successful_dial != last_successful_dial
        )
        """,
        [end, start, start, end]
    )
    return [i for sub in cur.fetchall() for i in sub]


# get_dangling_nodes gets the id of all dangling nodes between two timestamps.
# It takes an sql connection, the start time, the end time as the arguments, and
# returns the ids of all the dangling nodes.
def get_dangling_nodes(conn, start, end):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM sessions
        WHERE updated_at > %s AND updated_at < %s AND first_successful_dial != last_successful_dial
        """,
        [start, end]
    )
    return [i for sub in cur.fetchall() for i in sub]


# get_highly_dangling_nodes gets the id of all highly dangling nodes between two timestamps.
# It takes an sql connection, the start time, the end time, the number of sessions as the arguments, and
# returns the ids of all the highly dangling nodes.
def get_highly_dangling_nodes(conn, start, end, num):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT peer_id
        FROM sessions
        WHERE updated_at > %s AND updated_at < %s AND first_successful_dial != last_successful_dial
        GROUP BY peer_id
        HAVING COUNT(*) > %s
        """,
        [start, end, num]
    )
    return [i for sub in cur.fetchall() for i in sub]
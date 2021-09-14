import pytz


# get_nodes gets the id of on_nodes, off_nodes and dangling nodes between two timestamps.
# It takes an sql connection, the start time, the end time as the arguments, and
# returns the ids of the off_nodes, on_nodes, and dangling nodes.
def get_nodes(conn, start, end):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM connections
        WHERE dial_attempt > %s AND dial_attempt < %s AND peer_id NOT IN (
            SELECT DISTINCT peer_id
            FROM connections
            WHERE dial_attempt > %s AND dial_attempt < %s AND is_succeed IS true
        )
        """,
        [start, end, start, end]
    )
    off_nodes = cur.fetchall()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM connections
        WHERE dial_attempt > %s AND dial_attempt < %s AND peer_id NOT IN (
            SELECT DISTINCT peer_id
            FROM connections
            WHERE dial_attempt > %s AND dial_attempt < %s AND is_succeed IS false
        ) AND peer_id NOT IN (
            SELECT DISTINCT peer_id
            FROM sessions
            WHERE updated_at > %s AND updated_at < %s AND finish_reason = %s
        )
        """,
        [start, end, start, end, start, end, "maddr_reset"]
    )
    on_nodes = cur.fetchall()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM connections
        WHERE dial_attempt > %s AND dial_attempt < %s AND peer_id IN (
            SELECT DISTINCT peer_id
            FROM connections
            WHERE dial_attempt > %s AND dial_attempt < %s AND is_succeed IS true
        ) AND peer_id IN (
            SELECT DISTINCT peer_id
            FROM connections
            WHERE dial_attempt > %s AND dial_attempt < %s AND is_succeed IS false
            UNION
            SELECT DISTINCT peer_id
            FROM sessions
            WHERE updated_at > %s AND updated_at < %s AND finish_reason = %s
        )
        """,
        [start, end, start, end, start, end, start, end, "maddr_reset"]
    )
    dangling_nodes = cur.fetchall()
    return [i for sub in off_nodes for i in sub], [i for sub in on_nodes for i in sub], [i for sub in dangling_nodes for i in sub]

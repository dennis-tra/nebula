import pytz


# get_arrival_time calculates the inter arrival of given peers within a time range.
# It takes an sql connection, start time, the end time, the peer ids as arguments, and
# returns the reliabilities of these peer ids.
def get_arrival_time(conn, start, end, peer_ids):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    arrival = []
    for peer_id in peer_ids:
        cur.execute(
            """
            SELECT created_at, updated_at
            FROM sessions
            WHERE updated_at > %s AND updated_at < %s AND peer_id = %s
            ORDER BY updated_at ASC
            """,
            [start, end, peer_id]
        )
        sessions = cur.fetchall()
        while len(sessions) > 1:
            arrival.append((sessions[1][0] - sessions[0][1]).total_seconds())
            sessions = sessions[1:]
    return arrival

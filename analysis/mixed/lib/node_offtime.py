import pytz


# get_offtime calculates the offtime of given peers within a time range.
# It takes an sql connection, start time, the end time, the peer ids as arguments, and
# returns the reliabilities of these peer ids.
def get_offtime(conn, start, end, peer_ids):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    offtimes = []
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
        offtime = 0
        while len(sessions) > 0:
            if len(sessions) == 1:
                offtime += (end - sessions[0][1]).total_seconds()
            else:
                offtime += (sessions[1][0] - sessions[0][1]).total_seconds()
            sessions = sessions[1:]
        offtimes.append(offtime)
    return offtimes

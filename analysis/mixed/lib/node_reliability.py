import pytz


# get_reliability calculates the reliability of given peers within a time range.
# It takes an sql connection, start time, the end time, the peer ids as arguments, and
# returns the reliabilities of these peer ids.
def get_reliability(conn, start, end, peer_ids):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    reliabilities = []
    for peer_id in peer_ids:
        cur.execute(
            """
            SELECT created_at, min_duration
            FROM sessions
            WHERE updated_at > %s AND updated_at < %s AND peer_id = %s
            ORDER BY updated_at ASC
            """,
            [start, end, peer_id]
        )
        sessions = cur.fetchall()
        uptime = None
        total = 0
        if len(sessions) > 0:
            for session in sessions:
                if uptime is None:
                    if session[0] < start:
                        uptime = session[1] - (start - session[0])
                        total = end - start
                    else:
                        if session[1] is None:
                            uptime = (end - session[0])
                        else:
                            uptime = session[1]
                        total = end - session[0]
                else:
                    if session[1] is None:
                        uptime += (end - session[0])
                    else:
                        uptime += session[1]
            reliabilities.append(uptime / total)
        else:
            cur.execute(
                """
                SELECT *
                FROM sessions
                WHERE created_at < %s AND updated_at > %s AND peer_id = %s
                ORDER BY updated_at ASC
                """,
                [end, end, peer_id]
            )
            if len(cur.fetchall()) > 0:
                reliabilities.append(1.0)
            else:
                reliabilities.append(0.0)
    return reliabilities

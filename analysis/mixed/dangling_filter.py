import pytz


# get_dangling_counts gets the counts of the state switch of dangling nodes.
# It takes an sql connection, the start time, the end time, the peer ids as the arguments, and
# returns the state switch counts of there peer ids.
def get_dangling_counts(conn, start, end, peer_ids):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT peer_id, COUNT(*)
        FROM sessions
        WHERE updated_at > %s AND updated_at < %s AND peer_id IN (%s)
        GROUP BY peer_id
        """ % ("%s", "%s", ','.join(['%s'] * len(peer_ids))),
        (start, end) + tuple(peer_ids)
    )
    for id, count in cur.fetchall():
        res[id] = count
    return res

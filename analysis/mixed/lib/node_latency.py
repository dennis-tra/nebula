# get_latency gets the latency of given peer ids.
# It takes an sql connection, the start time, the end time, the peer ids as the arguments, and
# returns the latency of there peer ids.
def get_latency(conn, peer_ids):
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT peer_id, MAX(latency), MIN(latency), AVG(latency)
        FROM connections
        WHERE is_succeed = true AND peer_id IN (%s)
        GROUP BY peer_id
        """ % ','.join(['%s'] * len(peer_ids)),
        tuple(peer_ids)
    )
    for id, max, min, avg in cur.fetchall():
        res[id] = (max, min, avg)
    return res

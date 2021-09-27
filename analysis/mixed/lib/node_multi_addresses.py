# get_multi_addresses gets the multi addresses of the given peers.
# It takes an sql connection, the peer **database** ids as arguments, and
# returns the agent version info of these peer ids.
def get_multi_addresses(conn, peer_ids):
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT p.id, ma.maddr
        FROM peers p
        INNER JOIN peers_x_multi_addresses pxma on p.id = pxma.peer_id
        INNER JOIN multi_addresses ma on pxma.multi_address_id = ma.id
        WHERE p.id IN (%s)
        """ % ','.join(['%s'] * len(peer_ids)),
        tuple(peer_ids)
    )
    for id, maddr in cur.fetchall():
        res[id] = maddr
    return res

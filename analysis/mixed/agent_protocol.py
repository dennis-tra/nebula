# get_agent_protocol gets the agent protocol info of given peers.
# It takes an sql connection, the peer ids as arguments, and
# returns the agent protocol info of these peer ids.
def get_agent_protocol(conn, peer_ids):
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT id, protocol
        FROM peers
        WHERE id IN (%s)
        """ % ','.join(['%s'] * len(peer_ids)),
        tuple(peer_ids)
    )
    for id, protocol in cur.fetchall():
        res[id] = protocol.split(";")
    return res

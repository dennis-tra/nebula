# get_agent_version gets the agent version info of given peers.
# It takes an sql connection, the peer ids as arguments, and
# returns the agent version info of these peer ids.
def get_agent_version(conn, peer_ids):
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT id, agent_version
        FROM peers
        WHERE id IN (%s)
        """ % ','.join(['%s'] * len(peer_ids)),
        tuple(peer_ids)
    )
    for id, agent in cur.fetchall():
        res[id] = agent
    return res

# get_agent_version gets the agent version info of given peers.
# It takes an sql connection, the peer **database** ids as arguments, and
# returns the agent version info of these peer ids.
def get_agent_version(conn, peer_ids):
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT p.id, agent_version
        FROM peers p
        INNER JOIN agent_versions av on av.id = p.agent_version_id
        WHERE p.id IN (%s)
        """ % ','.join(['%s'] * len(peer_ids)),
        tuple(peer_ids)
    )
    for id, agent in cur.fetchall():
        res[id] = agent
    return res


# get_agent_version_multi_hash gets the agent version info of given peers.
# It takes an sql connection, the peer ID multi hashes as arguments, and
# returns the agent version info of these peer ids.
def get_agent_version_multi_hash(conn, peer_ids):
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT p.id, agent_version
        FROM peers p
        INNER JOIN agent_versions av on av.id = p.agent_version_id
        WHERE p.multi_hash IN (%s)
        """ % ','.join(['%s'] * len(peer_ids)),
        tuple(peer_ids)
    )
    for id, agent in cur.fetchall():
        res[id] = agent
    return res

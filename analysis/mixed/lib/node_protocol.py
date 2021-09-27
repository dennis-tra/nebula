# get_agent_protocol gets the agent protocol info of given peers.
# It takes an sql connection, the peer ids as arguments, and
# returns the agent protocol info of these peer ids.
def get_agent_protocol(conn, peer_ids):
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT sq.id, array_agg(prot.protocol)
        FROM protocols prot INNER JOIN (
            SELECT p.id, unnest(ps.protocol_ids) protocol_id
            FROM peers p INNER JOIN protocols_sets ps ON ps.id = p.protocols_set_id
            WHERE p.id IN (%s)
        ) AS sq ON sq.protocol_id = prot.id GROUP BY 1
        """ % ','.join(['%s'] * len(peer_ids)),
        tuple(peer_ids)
    )
    for id, protocols in cur.fetchall():
        res[id] = protocols
    return res

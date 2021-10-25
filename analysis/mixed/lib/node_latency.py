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


def update_peer_latency(conn, peer_id, location):
    cur = conn.cursor()
    cur.execute("""
    update latency set location=%s where peer_id=%s
    """, (location, peer_id))
    conn.commit()

def update_all_peers_latency(conn, locations):
    for peer_id, location in locations.items():
        update_peer_latency(conn, peer_id, location)

def get_latency_list(conn):
    cur = conn.cursor()
    cur.execute("""
    select location, EXTRACT(millisecond FROM avg(avg_latency))
    from latency 
    group by location
    order by 2
    """)
    location_list = []
    latency_ms_list = []
    for location, latency_ms in cur.fetchall():
        if not location:
            continue
        location_list.append(location)
        latency_ms_list.append(latency_ms)
        # print(latency_ms, location)
    return location_list, latency_ms_list
import geolocation_check
from multiaddr import Multiaddr
from which_cloud import WhichCloud
wc = WhichCloud()


# check_cloud gets the cloud info of given peers.
# It takes an sql connection, the peer ids as arguments, and
# returns the cloud info of these peer ids.
def check_cloud(conn, peer_ids):
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT id, multi_addresses
        FROM peers
        WHERE id IN (%s)
        """ % ','.join(['%s'] * len(peer_ids)),
        tuple(peer_ids)
    )
    for id, maddr_strs in cur.fetchall():
        found = False
        for maddr_str in maddr_strs:
            maddr = Multiaddr(maddr_str)
            try:
                address = geolocation_check.node_address(maddr)
                iso_code = get_clout(address)
                res[id] = iso_code
                found = True
                break
            except:
                pass
        if not found:
            res[id] = "unknown"
    return res


# Helper function, get cloud info from ip address.
def get_clout(ip):
    if wc.is_ip(ip) is None:
        return "unknown"
    return wc.is_ip(ip).name

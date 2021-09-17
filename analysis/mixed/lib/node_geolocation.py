import geoip2.database
from multiaddr import Multiaddr
import helper


# get_geolocation gets the geolocation info of given peers.
# It takes an sql connection, the peer ids as arguments, and
# returns the geolocation info of these peer ids.
def get_geolocation(conn, peer_ids):
    with geoip2.database.Reader("../geoip/GeoLite2/GeoLite2-Country.mmdb") as geoipreader:
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
                    address = helper.node_address(maddr)
                    iso_code = geoipreader.country(address).country.iso_code
                    if iso_code is None:
                        iso_code = "unknown"
                    res[id] = iso_code
                    found = True
                    break
                except:
                    pass
            if not found:
                res[id] = "unknown"
        return res

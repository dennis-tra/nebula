from plot_cloud import plot_cloud
from lib_db import DBClient

client = DBClient()
peer_ids = client.get_dangling_peer_ids()
results = client.query(
   f"""
    WITH cte AS (
        SELECT v.peer_id, unnest(mas.multi_address_ids) multi_address_id
        FROM visits v
                 INNER JOIN multi_addresses_sets mas on mas.id = v.multi_addresses_set_id
        WHERE v.created_at > date_trunc('week', NOW() - '1 week'::interval)
          AND v.created_at < date_trunc('week', NOW())
          AND v.peer_id IN ({",".join(str(x) for x in peer_ids)})
        GROUP BY v.peer_id, unnest(mas.multi_address_ids)
    )
    SELECT DISTINCT ia.address
    FROM multi_addresses ma
            INNER JOIN cte ON cte.multi_address_id = ma.id
            INNER JOIN multi_addresses_x_ip_addresses maxia on ma.id = maxia.multi_address_id
            INNER JOIN ip_addresses ia ON maxia.ip_address_id = ia.id
    """
)

plot_cloud(results, "Dangling", "dangling")

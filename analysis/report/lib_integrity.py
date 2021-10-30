from lib_db import DBClient

client = DBClient()

all_peer_ids = set(client.get_all_peer_ids())
online_peer_ids = set(client.get_online_peer_ids())
offline_peer_ids = set(client.get_offline_peer_ids())
all_entering_peer_ids = set(client.get_entering_peer_ids())
all_leaving_peer_ids = set(client.get_leaving_peer_ids())
dangling_peer_ids = set(client.get_dangling_peer_ids())

entering_peer_ids = all_entering_peer_ids.difference(all_leaving_peer_ids)
leaving_peer_ids = all_leaving_peer_ids.difference(all_entering_peer_ids)

assert online_peer_ids.issubset(all_peer_ids)
assert offline_peer_ids.issubset(all_peer_ids)
assert entering_peer_ids.issubset(all_peer_ids)
assert leaving_peer_ids.issubset(all_peer_ids)
assert dangling_peer_ids.issubset(all_peer_ids)

assert online_peer_ids.isdisjoint(offline_peer_ids)
assert online_peer_ids.isdisjoint(entering_peer_ids)
assert online_peer_ids.isdisjoint(leaving_peer_ids)
assert online_peer_ids.isdisjoint(dangling_peer_ids)
assert offline_peer_ids.isdisjoint(entering_peer_ids)

assert offline_peer_ids.isdisjoint(leaving_peer_ids)
assert offline_peer_ids.isdisjoint(dangling_peer_ids)

assert entering_peer_ids.isdisjoint(leaving_peer_ids)
assert entering_peer_ids.isdisjoint(dangling_peer_ids)

assert leaving_peer_ids.isdisjoint(dangling_peer_ids)


all_agent_versions = client.get_all_agent_versions()
all_peer_ids_by_agent_versions = set(client.get_peer_ids_for_agent_versions(all_agent_versions))

assert all_peer_ids_by_agent_versions.issubset(all_peer_ids)
assert entering_peer_ids.issubset(all_peer_ids_by_agent_versions)
assert online_peer_ids.issubset(all_peer_ids_by_agent_versions)
assert dangling_peer_ids.issubset(all_peer_ids_by_agent_versions)
assert offline_peer_ids.isdisjoint(all_peer_ids_by_agent_versions)

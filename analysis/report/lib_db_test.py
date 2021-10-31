import unittest

from lib_db import DBClient


class TestDBClient(unittest.TestCase):
    client: DBClient

    def setUp(self) -> None:
        self.client = DBClient()

    def test_integrity(self):
        all_peer_ids = set(self.client.get_all_peer_ids())

        online_peer_ids = set(self.client.get_online_peer_ids())
        self.assertTrue(online_peer_ids.issubset(all_peer_ids))

        offline_peer_ids = set(self.client.get_offline_peer_ids())
        self.assertTrue(offline_peer_ids.issubset(all_peer_ids))

        all_entering_peer_ids = set(self.client.get_all_entering_peer_ids())
        self.assertTrue(all_entering_peer_ids.issubset(all_peer_ids))
        self.assertTrue(all_entering_peer_ids.isdisjoint(online_peer_ids))
        self.assertTrue(all_entering_peer_ids.isdisjoint(offline_peer_ids))

        all_leaving_peer_ids = set(self.client.get_all_leaving_peer_ids())
        self.assertTrue(all_leaving_peer_ids.issubset(all_peer_ids))
        self.assertTrue(all_leaving_peer_ids.isdisjoint(online_peer_ids))
        self.assertTrue(all_leaving_peer_ids.isdisjoint(offline_peer_ids))

        # The following needn't be necessarily true but unlikely that it isn't
        self.assertTrue(len(all_entering_peer_ids.intersection(all_leaving_peer_ids)) > 0)

        only_entering_peer_ids = set(self.client.get_only_entering_peer_ids())
        self.assertTrue(only_entering_peer_ids.issubset(all_peer_ids))
        self.assertTrue(only_entering_peer_ids.isdisjoint(online_peer_ids))
        self.assertTrue(only_entering_peer_ids.isdisjoint(offline_peer_ids))
        self.assertTrue(only_entering_peer_ids.isdisjoint(all_leaving_peer_ids))
        self.assertTrue(only_entering_peer_ids.issubset(all_entering_peer_ids))

        only_leaving_peer_ids = set(self.client.get_only_leaving_peer_ids())
        self.assertTrue(only_leaving_peer_ids.issubset(all_peer_ids))
        self.assertTrue(only_leaving_peer_ids.isdisjoint(online_peer_ids))
        self.assertTrue(only_leaving_peer_ids.isdisjoint(offline_peer_ids))
        self.assertTrue(only_leaving_peer_ids.isdisjoint(all_entering_peer_ids))
        self.assertTrue(only_leaving_peer_ids.issubset(all_leaving_peer_ids))

        ephemeral_peer_ids = set(self.client.get_ephemeral_peer_ids())
        self.assertTrue(ephemeral_peer_ids.issubset(all_entering_peer_ids))
        self.assertTrue(ephemeral_peer_ids.issubset(all_leaving_peer_ids))

        dangling_peer_ids = set(self.client.get_dangling_peer_ids())
        self.assertTrue(dangling_peer_ids.issubset(all_peer_ids))
        self.assertTrue(dangling_peer_ids.isdisjoint(online_peer_ids))
        self.assertTrue(dangling_peer_ids.isdisjoint(offline_peer_ids))
        self.assertTrue(dangling_peer_ids.issubset(all_entering_peer_ids))
        self.assertTrue(dangling_peer_ids.issubset(all_leaving_peer_ids))

        oneoff_peer_ids = set(self.client.get_oneoff_peer_ids())
        self.assertTrue(oneoff_peer_ids.issubset(all_peer_ids))
        self.assertTrue(oneoff_peer_ids.isdisjoint(online_peer_ids))
        self.assertTrue(oneoff_peer_ids.isdisjoint(offline_peer_ids))
        self.assertTrue(oneoff_peer_ids.isdisjoint(dangling_peer_ids))
        self.assertTrue(oneoff_peer_ids.issubset(all_entering_peer_ids))
        self.assertTrue(oneoff_peer_ids.issubset(all_leaving_peer_ids))

        calculated_all_peer_ids = oneoff_peer_ids | online_peer_ids | offline_peer_ids | only_entering_peer_ids | only_leaving_peer_ids | dangling_peer_ids
        self.assertEqual(len(all_peer_ids), len(calculated_all_peer_ids))
        self.assertEqual(all_peer_ids, calculated_all_peer_ids)

    def test_get_all_peer_ids_for_all_agent_versions(self):
        all_agent_versions = self.client.get_all_agent_versions()
        all_peer_ids_by_all_agent_versions = set(self.client.get_peer_ids_for_agent_versions(all_agent_versions))

        online_peer_ids = set(self.client.get_online_peer_ids())
        all_entering_peer_ids = set(self.client.get_all_entering_peer_ids())
        dangling_peer_ids = set(self.client.get_dangling_peer_ids())

        self.assertTrue(online_peer_ids.issubset(all_peer_ids_by_all_agent_versions))
        self.assertTrue(all_entering_peer_ids.issubset(all_peer_ids_by_all_agent_versions))
        self.assertTrue(dangling_peer_ids.issubset(all_peer_ids_by_all_agent_versions))

        # Now there can be nodes that started their session before
        # the beginning of the time interval, were then "crawlable" (we
        # could extract the agent version) and then left.
        left_peer_ids = all_peer_ids_by_all_agent_versions - online_peer_ids - all_entering_peer_ids - dangling_peer_ids
        only_leaving_peer_ids = set(self.client.get_only_leaving_peer_ids())
        self.assertTrue(left_peer_ids.issubset(only_leaving_peer_ids))

        # TODO: there is a minor bug in the time calculation of session start/ends. When that's fixed:
        # self.assertEqual(left_peer_ids, only_leaving_peer_ids)

    def test_agent_version_queries(self):
        agent_version_distribution = self.client.get_agent_versions_distribution()
        agent_version = agent_version_distribution[0][0]
        agent_version_count = agent_version_distribution[0][1]
        peer_ids_by_agent_version = self.client.get_peer_ids_for_agent_versions([agent_version])
        self.assertEqual(agent_version_count, len(peer_ids_by_agent_version))

        agent_versions_for_peer_ids = self.client.get_agent_versions_for_peer_ids(peer_ids_by_agent_version)
        self.assertEqual(agent_versions_for_peer_ids[0][1],
                         agent_version_count)  # we only queried for peers with one agent

    def test_geo_integrity(self):
        import pandas as pd

        all_peer_ids = set(self.client.get_all_peer_ids())

        no_public_ip_peer_ids = set(self.client.get_no_public_ip_peer_ids())
        self.assertTrue(no_public_ip_peer_ids.issubset(all_peer_ids))

        countries = self.client.get_countries()
        countries_peer_ids = set(pd.DataFrame(countries, columns=["peer_id", "country"])["peer_id"].unique())
        self.assertTrue(countries_peer_ids.issubset(all_peer_ids))
        self.assertTrue(countries_peer_ids.isdisjoint(no_public_ip_peer_ids))

        countries_with_relays = self.client.get_countries_with_relays()
        countries_with_relays_peer_ids = set(
            pd.DataFrame(countries_with_relays, columns=["peer_id", "country"])["peer_id"].unique())
        self.assertTrue(countries_with_relays_peer_ids.issubset(all_peer_ids))
        self.assertTrue(countries_with_relays_peer_ids.isdisjoint(no_public_ip_peer_ids))
        self.assertTrue(countries_peer_ids.issubset(countries_with_relays_peer_ids))

        unresolved_peer_ids = set(self.client.get_unresolved_peer_ids())
        self.assertTrue(unresolved_peer_ids.issubset(all_peer_ids))
        self.assertTrue(unresolved_peer_ids.isdisjoint(no_public_ip_peer_ids))
        self.assertTrue(unresolved_peer_ids.isdisjoint(countries_peer_ids))
        self.assertTrue(unresolved_peer_ids.isdisjoint(countries_with_relays_peer_ids))

        calculated_all = no_public_ip_peer_ids | countries_peer_ids | countries_with_relays_peer_ids | unresolved_peer_ids

        self.assertEqual(all_peer_ids, calculated_all)

    def test_flatten(self):
        flattened = DBClient._DBClient__flatten([(1,), (2,)])
        self.assertListEqual(flattened, [1, 2])


if __name__ == '__main__':
    unittest.main()

import toml
import os
import json
import psycopg2
import datetime
import hashlib
import pandas as pd

from enum import Enum
from typing import TypeVar

T = TypeVar('T')

calendar = (datetime.date.today() - datetime.timedelta(weeks=1)).isocalendar()
calendar_week = calendar.week
year = calendar.year


class NodeClassification(Enum):
    OFFLINE = "offline"
    ONEOFF = "oneoff"
    DANGLING = "dangling"
    ONLINE = "online"
    ENTERED = "entered"
    LEFT = "left"


class IPResolutionClassification(Enum):
    RESOLVED = "resolved"
    UNRESOLVED = "unresolved"
    NO_PUBLIC_IP = "no public ip"
    RELAY_ONLY = "relay_only"


def cache():
    """
    cache is a decorator that first checks the existence of a cache file before
    resorting to actually query the database. It takes the cache file name
    as a parameter. The cache files are scope by the calendar week as
    all queries only look for the most recent completed week
    """

    def decorator(func):
        def wrapper(*args, **kwargs):

            filename = func.__name__

            if not os.path.isdir(".cache"):
                os.mkdir(".cache")
            client: DBClient = args[0]
            hash_str = str(calendar_week) + \
                       str(client.start) + \
                       str(client.end) + \
                       str(client.config) + \
                       str(args[1:])
            digest = hashlib.sha256(str.encode(hash_str)).hexdigest()
            cache_file = f'.cache/{filename}-{digest}.json'

            if os.path.isfile(cache_file):
                print(f"Using cache file {cache_file} for {filename}...")
                with open(cache_file, 'r') as f:
                    return json.load(f)

            result = func(*args, **kwargs)

            with open(cache_file, 'w') as f:
                json.dump(result, f)

            return result

        return wrapper

    return decorator


class DBClient:
    config = None
    conn = None
    start = "date_trunc('week', NOW() - '1 week'::interval)"
    end = "date_trunc('week', NOW())"
    node_classification_funcs = {}

    @staticmethod
    def __flatten(result: list[tuple[T]]) -> list[T]:
        """
        flatten turns a list of 1d tuples like [(1,), (2,)] into
        a list like [1, 2]
        """
        return [i for sub in result for i in sub]

    @staticmethod
    def fmt_list(items: list[T]) -> str:
        return ",".join(str(elem) for elem in items)

    def __init__(self):
        print("Initializing database client...")

        self.config = toml.load("./db.toml")['psql']
        self.conn = psycopg2.connect(
            host=self.config['host'],
            port=self.config['port'],
            database=self.config['database'],
            user=self.config['user'],
            password=self.config['password'],
        )

        self.node_classification_funcs = {
            NodeClassification.OFFLINE: self.get_offline_peer_ids,
            NodeClassification.ONEOFF: self.get_oneoff_peer_ids,
            NodeClassification.DANGLING: self.get_dangling_peer_ids,
            NodeClassification.ONLINE: self.get_online_peer_ids,
            NodeClassification.ENTERED: self.get_only_entering_peer_ids,
            NodeClassification.LEFT: self.get_only_leaving_peer_ids,
        }

    @cache()
    def query(self, query):
        print("Running custom query...")
        cur = self.conn.cursor()
        cur.execute(query)
        return cur.fetchall()

    @cache()
    def get_all_peer_ids(self) -> list[int]:  #
        """
        get_all_peer_ids returns the set of **database** peer IDs
        that were visited during the specified time interval.
        """
        print("Getting all visited database peer IDs in specified time interval...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT DISTINCT peer_id
            FROM visits
            WHERE created_at > {self.start}
              AND created_at < {self.end}
            """
        )
        return DBClient.__flatten(cur.fetchall())

    @cache()
    def get_online_peer_ids(self) -> list[int]:  #
        """
        get_online_peer_ids returns the set of **database** peer IDs of
        all nodes that haven't been seen offline in the specified time
        interval. They were seen online the whole time.
        """
        print("Getting online database peer IDs in specified time interval...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT DISTINCT peer_id
            FROM sessions
            WHERE first_successful_dial < {self.start}
              AND (first_failed_dial > {self.end} OR finished = false)
            """
        )
        return DBClient.__flatten(cur.fetchall())

    @cache()
    def get_offline_peer_ids(self) -> list[int]:  #
        """
        get_offline_peer_ids returns the set of **database** peer IDs of
        all nodes that were visited in the specified time interval but
        never have been seen online (never dialable). E.g., they could
        have been found in the DHT but were never reachable or an active
        session from before the time interval overlaps into this time interval
        but the associated peer could never be contact in this time interval.
        """
        print("Getting offline database peer IDs in specified time interval...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT DISTINCT v.peer_id
            FROM visits v
            WHERE created_at > {self.start}
              AND created_at < {self.end}
              AND v.peer_id NOT IN (
                -- This subquery fetches all peers that have been
                -- seen online in the given time interval. We check if there is at least
                -- one visit without an error in the given time interval.
                SELECT DISTINCT v.peer_id
                FROM visits v
                         LEFT JOIN sessions s ON v.session_id = s.id
                WHERE v.created_at > {self.start}
                  AND v.created_at < {self.end}
                  AND v.error IS NULL
            )
            """
        )
        return DBClient.__flatten(cur.fetchall())

    @cache()
    def get_all_entering_peer_ids(self) -> list[int]:  #
        """
        get_all_entering_peer_ids returns the set **database** peer IDs of
        all nodes that started at least one new session during the
        specified time interval. This set can overlap with the set of
        leaving peer IDs from get_all_leaving_peer_ids.
        """
        print("Getting entering database peer IDs in specified time interval...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT DISTINCT peer_id
            FROM sessions
            WHERE first_successful_dial > {self.start}
              AND first_successful_dial < {self.end}
            """
        )
        return DBClient.__flatten(cur.fetchall())

    def get_only_entering_peer_ids(self) -> list[int]:  #
        """
        get_only_entering_peer_ids returns the set **database** peer IDs of
        all nodes that were offline at the beginning of the specified time
        interval, then started only one new session and then didn't go
        offline until the end of the time interval.
        """
        print("Getting only entering database peer IDs in specified time interval...")
        return list(set(self.get_all_entering_peer_ids()).difference(set(self.get_all_leaving_peer_ids())))

    @cache()
    def get_all_leaving_peer_ids(self) -> list[int]:  #
        """
        get_all_leaving_peer_ids returns the set of **database** peer IDs of
        all nodes that ended a session at least once during the specified
        time interval. This set can overlap with the set of
        leaving peer IDs from get_all_entering_peer_ids.
        """
        print("Getting leaving database peer IDs in specified time interval...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT DISTINCT peer_id
            FROM sessions
            WHERE first_failed_dial < {self.end}
              AND first_failed_dial > {self.start}
              AND last_successful_dial > {self.start}
            """
        )
        return DBClient.__flatten(cur.fetchall())

    def get_only_leaving_peer_ids(self) -> list[int]:  #
        """
        get_only_leaving_peer_ids returns the set **database** peer IDs of
        all nodes that were online at the beginning of the specified time
        interval, then ended their session (went offline) and then didn't come
        back online until the end of the time interval.
        """
        print("Getting only leaving database peer IDs in specified time interval...")
        return list(set(self.get_all_leaving_peer_ids()).difference(set(self.get_all_entering_peer_ids())))

    def get_ephemeral_peer_ids(self) -> list[int]:  #
        """
        get_ephemeral_peer_ids returns the set of **database** peer IDs that
        entered the network but also left the network in the specified time
        interval. This may have happened multiple times.
        """
        print("Getting ephemeral database peer IDs in specified time interval...")
        return list(set(self.get_all_entering_peer_ids()).intersection(set(self.get_all_leaving_peer_ids())))

    def get_dangling_peer_ids(self) -> list[int]:  #
        """
        get_dangling_peer_ids returns the set of **database** peer IDs of
        all nodes that ended their online session during the specified time
        interval and also came online again (possibly multiple times).
        """
        print("Getting dangling database peer IDs in specified time interval...")
        return list(set(self.get_ephemeral_peer_ids()) - set(self.get_oneoff_peer_ids()))

    @cache()
    def get_oneoff_peer_ids(self) -> list[int]:  #
        """
        get_oneoff_peer_ids returns the set of **database** peer IDs that
        are associated with only one session in the specified time interval.
        This only includes sessions that completely lie within this interval,
        e.g., sessions that started before the beginning of the interval and
        ended within are excluded.
        """
        print("Getting one off database peer IDs in specified time interval...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT peer_id
            FROM sessions
            WHERE created_at < {self.end}
              AND updated_at > {self.start}
              AND peer_id IN ({self.fmt_list(self.get_ephemeral_peer_ids())})
            GROUP BY peer_id
            HAVING count(id) = 1
            """
        )
        return DBClient.__flatten(cur.fetchall())

    def get_all_agent_versions(self) -> list[str]:  #
        """
        get_all_agent_versions returns all ever discovered agent versions
        ordered by their discovery date.
        """
        print("Getting all agent versions...")
        cur = self.conn.cursor()
        cur.execute("SELECT agent_version FROM agent_versions ORDER BY created_at")
        return DBClient.__flatten(cur.fetchall())

    @cache()
    def get_peer_ids_for_agent_versions(self, agent_versions: list[str]):  #
        """
        get_peer_ids_for_agent_versions returns the set of **database** peer IDs
        that reported at least one agent version in the given list.
        """
        print(f"Getting database peer IDs for {agent_versions} agent versions...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT DISTINCT v.peer_id
            FROM visits v
                     INNER JOIN agent_versions av on av.id = v.agent_version_id
            WHERE v.created_at > {self.start}
              AND v.created_at < {self.end}
              AND v.type = 'crawl'
              AND v.error IS NULL
              AND av.agent_version LIKE ANY (array[{",".join(f"'%{av}%'" for av in agent_versions)}])
            """
        )
        return DBClient.__flatten(cur.fetchall())

    @cache()
    def get_agent_versions_distribution(self) -> list[tuple[str, int]]:  #
        """
        get_agent_versions_distribution returns all agent versions with
        a count of peers that were discovered with such an agent version.
        """
        print("Getting agent versions distribution...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT av.agent_version, count(DISTINCT peer_id) "count"
            FROM visits v
            INNER JOIN agent_versions av on av.id = v.agent_version_id
            WHERE v.created_at > {self.start}
              AND v.created_at < {self.end}
              AND v.type = 'crawl'
              AND v.error IS NULL
            GROUP BY av.agent_version
            ORDER BY count DESC
            """
        )
        return cur.fetchall()

    @cache()
    def get_agent_versions_for_peer_ids(self, peer_ids: list[int]):  #
        """
        get_agent_versions_for_peer_ids returns all agent versions with
        a count of peers that were discovered with such an agent version
        from the list of given peer_ids.
        """
        print(f"Getting agent versions for {len(peer_ids)} peers...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT av.agent_version, count(DISTINCT peer_id) "count"
            FROM visits v
            INNER JOIN agent_versions av on av.id = v.agent_version_id
            WHERE v.created_at > {self.start}
              AND v.created_at < {self.end}
              AND v.type = 'crawl'
              AND v.error IS NULL
              AND v.peer_id IN ({self.fmt_list(peer_ids)})
            GROUP BY av.agent_version
            ORDER BY count DESC
            """
        )
        return cur.fetchall()

    @cache()
    def get_crawls(self):  #
        """
        get_crawls returns information of all crawls in the
        specified time interval.
        """
        print("Getting general crawl information...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT extract(epoch from started_at), crawled_peers, dialable_peers, undialable_peers
            FROM crawls c
            WHERE c.created_at > {self.start}
              AND c.created_at < {self.end}
            """
        )
        return cur.fetchall()

    @cache()
    def get_crawl_properties(self, threshold=10):  #
        """
        get_crawl_properties returns agent version distributions
        of the crawls in the specified time interval.
        """
        print("Getting crawl properties...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT cp.crawl_id, EXTRACT('epoch' FROM c.started_at) started_at, av.agent_version, cp.count
            FROM crawl_properties cp 
                INNER JOIN agent_versions av ON cp.agent_version_id = av.id
                INNER JOIN crawls c ON cp.crawl_id = c.id
            WHERE cp.created_at > {self.start}
              AND cp.created_at < {self.end}
              AND cp.count > {threshold}
            """
        )
        return cur.fetchall()

    @cache()
    def get_crawl_visit_durations(self):  #
        """
        get_crawl_visit_durations returns all durations for connecting
        and crawling peers in the specified time interval.
        """
        print("Getting crawl visit durations...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT
                EXTRACT('epoch' FROM v.connect_duration), 
                EXTRACT('epoch' FROM v.crawl_duration)
            FROM visits v
            WHERE v.created_at > {self.start}
              AND v.created_at < {self.end}
              AND v.type = 'crawl'
              AND v.error IS NULL
            """
        )
        return cur.fetchall()

    @cache()
    def get_dial_visit_durations(self):  #
        """
        get_dial_visit_durations returns all durations for
        dialing peers in the specified time interval.
        """
        print("Getting dial visit durations...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT
                EXTRACT('epoch' FROM v.dial_duration)
            FROM visits v
            WHERE v.created_at > {self.start}
              AND v.created_at < {self.end}
              AND v.type = 'dial'
              AND v.error IS NULL
            """
        )
        return cur.fetchall()

    @cache()
    def get_node_uptime(self):
        """
        get_node_uptime gets the session uptimes of the last completed week.
        It returns the a list containing the uptime in seconds.
        """
        print("Getting node uptimes...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT EXTRACT(EPOCH FROM min_duration), av.agent_version
            FROM sessions s
            INNER JOIN peers p on s.peer_id = p.id
            INNER JOIN agent_versions av on p.agent_version_id = av.id
            WHERE s.created_at < {self.end}
              AND s.updated_at > {self.start}
            """
        )
        return cur.fetchall()

    @cache()
    def get_inter_arrival_time(self, peer_ids):
        """
        get_inter_arrival_time returns the times between two sessions of the
        same peer.
        """
        print(f"Getting inter arrival times for {len(peer_ids)} peers...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT s1.id,
                   s1.peer_id,
                   EXTRACT('epoch' FROM MIN(s2.created_at) - s1.created_at) AS diff_in_s
            FROM sessions s1
                     LEFT JOIN sessions s2 ON s1.peer_id = s2.peer_id AND s1.created_at < s2.created_at
            WHERE s1.updated_at > {self.start}
              AND s1.created_at < {self.end}
              AND s2.created_at IS NOT NULL AND s1.peer_id IN ({self.fmt_list(peer_ids)})
            GROUP BY s1.id, s1.peer_id
            ORDER BY s1.created_at;
            """
        )
        return cur.fetchall()

    @cache()
    def get_ip_addresses_for_peer_ids(self, peer_ids):
        print("Getting ip addresses for peer IDs...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            WITH cte AS (
                SELECT v.peer_id, unnest(mas.multi_address_ids) multi_address_id
                FROM visits v
                         INNER JOIN multi_addresses_sets mas on mas.id = v.multi_addresses_set_id
                WHERE v.created_at > {self.start}
                  AND v.created_at < {self.end}
                  AND v.peer_id IN ({self.fmt_list(peer_ids)})
                GROUP BY v.peer_id, unnest(mas.multi_address_ids)
            )
            SELECT DISTINCT ia.address
            FROM multi_addresses ma
                    INNER JOIN cte ON cte.multi_address_id = ma.id
                    INNER JOIN multi_addresses_x_ip_addresses maxia on ma.id = maxia.multi_address_id
                    INNER JOIN ip_addresses ia ON maxia.ip_address_id = ia.id
            """
        )
        return cur.fetchall()

    @cache()
    def get_no_public_ip_peer_ids(self) -> list[int]:  #
        """
        get_no_public_ip_peer_ids returns the set of **database** peer IDs that
        didn't have a public IP address throughout the specified time interval.
        Still don't know how this is possible...
        """
        print("Getting database peer IDs with no public IP in specified time interval...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            SELECT DISTINCT peer_id 
            FROM visits v
            WHERE v.created_at > {self.start}
              AND v.created_at < {self.end}
              AND v.multi_addresses_set_id IS NULL
              AND peer_id NOT IN (
                SELECT DISTINCT peer_id from visits v
                WHERE v.created_at > {self.start}
                  AND v.created_at < {self.end}
                  AND v.multi_addresses_set_id IS NOT NULL
                );
            """
        )
        return DBClient.__flatten(cur.fetchall())

    @cache()
    def get_unresolved_peer_ids(self) -> list[int]:  #
        """
        get_unresolved_peer_ids returns the set of **database** peer IDs that
        had a public IP address but weren't or couldn't be resolved to an IP address.
        """
        print("Getting unresolved database peer IDs in specified time interval...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            WITH peer_maddrs AS (
                SELECT v.peer_id, unnest(mas.multi_address_ids) multi_address_id
                FROM visits v
                         INNER JOIN multi_addresses_sets mas on mas.id = v.multi_addresses_set_id
                WHERE v.created_at > {self.start}
                  AND v.created_at < {self.end}
                GROUP BY v.peer_id, unnest(mas.multi_address_ids)
            ), peer_maddrs_resolved AS (
                SELECT pm.peer_id, maxia.multi_address_id
                FROM peer_maddrs pm
                         LEFT JOIN multi_addresses_x_ip_addresses maxia on pm.multi_address_id = maxia.multi_address_id
                GROUP BY pm.peer_id, maxia.multi_address_id
            )
            SELECT DISTINCT peer_id
            FROM peer_maddrs_resolved
            WHERE peer_id NOT IN (
                SELECT DISTINCT peer_id
                FROM peer_maddrs_resolved
                WHERE multi_address_id IS NOT NULL
            )
            """
        )
        return DBClient.__flatten(cur.fetchall())

    @cache()
    def get_countries(self):
        """
        get_countries returns a list of peer IDs and their corresponding countries if they happened
        to have an IP address in in the specified time interval. Each peer ID can be associated
        to many multi addresses which in turn are associated to many IP addresses. Often, the
        number of associated ip addresses is smaller than the number of multi addresses as many
        multi addresses only differ in the protocol being used. Even if there are multiple IP
        addresses associated the number of different countries they are belonging to is often
        smaller than the number of IP addresses as well.
        This method EXCLUDES peers that were only reachable via circuit relays.
        """
        print("Getting countries information (without relays)...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            WITH peer_maddrs AS (
                SELECT v.peer_id, unnest(mas.multi_address_ids) multi_address_id
                FROM visits v
                         INNER JOIN multi_addresses_sets mas on mas.id = v.multi_addresses_set_id
                WHERE v.created_at > {self.start}
                  AND v.created_at < {self.end}
                GROUP BY v.peer_id, unnest(mas.multi_address_ids)
            )
            SELECT pm.peer_id, ia.country
            FROM multi_addresses ma
                     INNER JOIN peer_maddrs pm ON pm.multi_address_id = ma.id
                     INNER JOIN multi_addresses_x_ip_addresses maxia on pm.multi_address_id = maxia.multi_address_id
                     INNER JOIN ip_addresses ia on maxia.ip_address_id = ia.id
            WHERE ma.maddr NOT LIKE '%p2p-circuit%'
            GROUP BY pm.peer_id, ia.country
            """
        )
        return cur.fetchall()

    @cache()
    def get_countries_with_relays(self):
        """
        get_countries_with_relays returns a list of peer IDs and their corresponding countries if they happened
        to have an IP address in in the specified time interval. Each peer ID can be associated
        to many multi addresses which in turn are associated to many IP addresses. Often, the
        number of associated ip addresses is smaller than the number of multi addresses as many
        multi addresses only differ in the protocol being used. Even if there are multiple IP
        addresses associated the number of different countries they are belonging to is often
        smaller than the number of IP addresses as well.
        This method INCLUDES peers that were only reachable via circuit relays.
        """
        print("Getting countries information (with relays)...")
        cur = self.conn.cursor()
        cur.execute(
            f"""
            WITH peer_maddrs AS (
                SELECT v.peer_id, unnest(mas.multi_address_ids) multi_address_id
                FROM visits v
                         INNER JOIN multi_addresses_sets mas on mas.id = v.multi_addresses_set_id
                WHERE v.created_at > {self.start}
                  AND v.created_at < {self.end}
                GROUP BY v.peer_id, unnest(mas.multi_address_ids)
            )
            SELECT pm.peer_id, ia.country
            FROM multi_addresses ma
                     INNER JOIN peer_maddrs pm ON pm.multi_address_id = ma.id
                     INNER JOIN multi_addresses_x_ip_addresses maxia on pm.multi_address_id = maxia.multi_address_id
                     INNER JOIN ip_addresses ia on maxia.ip_address_id = ia.id
            GROUP BY pm.peer_id, ia.country
            """
        )
        return cur.fetchall()

    def get_country_distribution_for_peer_ids(self, peer_ids):
        data = pd.DataFrame(self.get_countries(), columns=["peer_id", "country"])
        data = data[data["peer_id"].isin(peer_ids)]
        data = data.groupby(by="country", as_index=False).count().sort_values('peer_id', ascending=False)
        return data.rename(columns={'country': 'Country', 'peer_id': 'Count'})

import toml
import os
import json
import psycopg2
import datetime
import hashlib

calendar_week = (datetime.date.today() - datetime.timedelta(weeks=1)).isocalendar().week


def cache(filename: str):
    """
    cache is a decorator that first checks the existence of a cache file before
    resorting to actually query the database. It takes the cache file name
    as a parameter. The cache files are scope by the calendar week as
    all queries only look for the most recent completed week
    """

    def decorator(func):
        def wrapper(*args, **kwargs):
            cache_file = f'.cache/{filename}-{calendar_week}.json'
            if len(args) == 2:
                digest = hashlib.sha256(str.encode(str(args[1]))).hexdigest()
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

    @cache("query")
    def query(self, query):
        print("Running custom query...")
        cur = self.conn.cursor()
        cur.execute(query)
        return cur.fetchall()

    @cache("get_visited_peers_agent_versions")
    def get_visited_peers_agent_versions(self):
        """
        get_visited_peers_agent_versions gets the agent version
        counts of the peers that were visited during the last
        completed week.
        """
        print("Getting agent versions for visited peers...")
        cur = self.conn.cursor()
        cur.execute(
            """
            SELECT av.agent_version, count(DISTINCT peer_id) "count"
            FROM visits v
            INNER JOIN agent_versions av on av.id = v.agent_version_id
            WHERE v.created_at > date_trunc('week', NOW() - '1 week'::interval)
              AND v.created_at < date_trunc('week', NOW())
              AND v.type = 'crawl'
              AND v.error IS NULL
            GROUP BY av.agent_version
            ORDER BY count DESC
            """
        )
        return cur.fetchall()

    @cache("get_agent_versions_for_peer_ids")
    def get_agent_versions_for_peer_ids(self, peer_ids):
        """
        get_visited_peers_agent_versions gets the agent version
        counts of the peers that were visited during the last
        completed week.
        """
        print(f"Getting agent versions for {len(peer_ids)} peers...")
        cur = self.conn.cursor()
        cur.execute(
            """
            SELECT av.agent_version, count(DISTINCT peer_id) "count"
            FROM visits v
            INNER JOIN agent_versions av on av.id = v.agent_version_id
            WHERE v.created_at > date_trunc('week', NOW() - '1 week'::interval)
              AND v.created_at < date_trunc('week', NOW())
              AND v.type = 'crawl'
              AND v.error IS NULL
              AND v.peer_id IN (%s)
            GROUP BY av.agent_version
            ORDER BY count DESC
            """ % ','.join(['%s'] * len(peer_ids)),
            tuple(peer_ids)
        )
        return cur.fetchall()

    @cache("get_node_uptime")
    def get_node_uptime(self):
        """
        get_node_uptime gets the session uptimes of the last completed week.
        It returns the a list containing the uptime in seconds.
        """
        print("Getting node uptimes...")
        cur = self.conn.cursor()
        cur.execute(
            """
            SELECT EXTRACT(EPOCH FROM min_duration), av.agent_version
            FROM sessions s
            INNER JOIN peers p on s.peer_id = p.id
            INNER JOIN agent_versions av on p.agent_version_id = av.id
            WHERE s.created_at < date_trunc('week', NOW())
              AND s.updated_at > date_trunc('week', NOW() - '1 week'::interval)
            """
        )
        return cur.fetchall()

    @cache("get_all_peer_ids")
    def get_all_peer_ids(self):
        """
        get_all_peer_ids returns the set of peer IDs that were
        visited during the most recent complete week (not the current
        one). It returns a list of distinct **database** peer IDs.
        """
        print("Getting database peer IDs from last week...")
        cur = self.conn.cursor()
        cur.execute(
            """
            SELECT DISTINCT peer_id
            FROM visits
            WHERE created_at > date_trunc('week', NOW() - '1 week'::interval)
              AND created_at < date_trunc('week', NOW())
            """
        )
        return [i for sub in cur.fetchall() for i in sub]

    @cache("get_online_peer_ids")
    def get_online_peer_ids(self):
        """
        get_online_peer_ids gets the **database** ids of all nodes that haven't been seen offline in the last
        completed week. They were seen online the whole time.
        """
        print("Getting online database peer IDs from last week...")
        cur = self.conn.cursor()
        cur.execute(
            """
            SELECT DISTINCT peer_id
            FROM sessions
            WHERE first_successful_dial < date_trunc('week', NOW() - '1 week'::interval)
              AND (first_failed_dial > date_trunc('week', NOW()) OR finished = false)
            """
        )
        return [i for sub in cur.fetchall() for i in sub]

    @cache("get_offline_peer_ids")
    def get_offline_peer_ids(self):
        """
        get_offline_peer_ids gets the **database** ids of all nodes that haven't been seen online in the last
        completed week. They were found in the DHT but were never reachable the whole time.
        """
        print("Getting offline database peer IDs from last week...")
        cur = self.conn.cursor()
        cur.execute(
            """
            SELECT DISTINCT v.peer_id
            FROM visits v
            WHERE created_at > date_trunc('week', NOW() - '1 week'::interval)
              AND created_at < date_trunc('week', NOW())
              AND v.peer_id NOT IN (
                -- This subquery fetches all peers that have been
                -- seen online in the given time interval. We check if there is at least
                -- one visit without an error in the given time interval. Alternatively
                -- we check if there is a visit with an associated session (also an
                -- indication that the peer was online, but only if the first failed
                -- dial of that peer was in the given time interval.
                SELECT DISTINCT v.peer_id
                FROM visits v
                         LEFT JOIN sessions s ON v.session_id = s.id
                WHERE v.created_at > date_trunc('week', NOW() - '1 week'::interval)
                  AND v.created_at < date_trunc('week', NOW())
                  AND (v.error IS NULL OR
                       (v.session_id IS NOT NULL AND s.first_failed_dial > date_trunc('week', NOW() - '1 week'::interval)))
            )
            """
        )
        return [i for sub in cur.fetchall() for i in sub]

    @cache("get_entering_peer_ids")
    def get_entering_peer_ids(self):
        """
        get_entering_peer_ids gets the **database** ids of all nodes that started at least
        one new session during the last completed week.
        """
        print("Getting entering database peer IDs from last week...")
        cur = self.conn.cursor()
        cur.execute(
            """
            SELECT DISTINCT peer_id
            FROM sessions
            WHERE first_successful_dial > date_trunc('week', NOW() - '1 week'::interval)
              AND first_successful_dial < date_trunc('week', NOW())
            """
        )
        return [i for sub in cur.fetchall() for i in sub]

    @cache("get_leaving_peer_ids")
    def get_leaving_peer_ids(self):
        """
        get_leaving_peer_ids gets the **database** ids of all nodes that ended a session
        at least once during the last completed week.
        """
        print("Getting leaving database peer IDs from last week...")
        cur = self.conn.cursor()
        cur.execute(
            """
            SELECT DISTINCT peer_id
            FROM sessions
            WHERE first_failed_dial < date_trunc('week', NOW())
              AND first_failed_dial > date_trunc('week', NOW() - '1 week'::interval)
            """
        )
        return [i for sub in cur.fetchall() for i in sub]

    @cache("get_dangling_peer_ids")
    def get_dangling_peer_ids(self):
        """
        get_dangling_peer_ids gets the **database** ids of all nodes that ended their online session
        during the last completed week and also came online again (possibly multiple times).
        """
        all_entering_peer_ids = set(self.get_entering_peer_ids())
        all_leaving_peer_ids = set(self.get_leaving_peer_ids())
        return list(all_entering_peer_ids.intersection(all_leaving_peer_ids))

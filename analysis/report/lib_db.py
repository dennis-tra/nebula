import toml
import os
import json
import psycopg2
import datetime


class DBClient:
    config = None
    conn = None
    calendar_week = (datetime.date.today() - datetime.timedelta(weeks=1)).isocalendar().week

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

    def get_last_weeks_peer_ids(self):
        """
        get_last_weeks_peer_ids returns the set of peer IDs that were
        visited during the most recent complete week (not the current
        one). It returns a list of distinct **database** peer IDs.
        """
        print("Getting database peer IDs from last week...")
        cache_file = '.cache/peer_ids-%s.json' % self.calendar_week
        if os.path.isfile(cache_file):
            print("Using peer ID cache...")
            with open(cache_file, 'r') as f:
                return json.load(f)

        cur = self.conn.cursor()
        cur.execute(
            """
            SELECT DISTINCT peer_id
            FROM visits
            WHERE created_at > date_trunc('week', NOW() - '1 week'::interval)
              AND created_at < date_trunc('week', NOW())
              AND error IS NULL
            """
        )
        result = [i for sub in cur.fetchall() for i in sub]

        with open(cache_file, 'w') as f:
            json.dump(result, f)

        return result

    def get_visited_peers_agent_versions(self):
        """
        get_visited_peers_agent_versions gets the agent version
        counts of the peers that were visited during the last
        completed week.
        """
        print("Getting agent versions for visited peers...")
        cache_file = '.cache/get_visited_peers_agent_versions-%s.json' % self.calendar_week
        if os.path.isfile(cache_file):
            print("Using cache...")
            with open(cache_file, 'r') as f:
                return json.load(f)
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
        result = cur.fetchall()
        with open(cache_file, 'w') as f:
            json.dump(result, f)
        return result

    def get_node_uptime(self):
        """
        get_node_uptime gets the session uptimes of the last completed week.
        It returns the a list containing the uptime in seconds.
        """
        print("Getting node uptimes...")
        cache_file = '.cache/get_node_uptime-%s.json' % self.calendar_week
        if os.path.isfile(cache_file):
            print("Using cache...")
            with open(cache_file, 'r') as f:
                return json.load(f)

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
        result = cur.fetchall()
        with open(cache_file, 'w') as f:
            json.dump(result, f)
        return result

    def get_on_nodes(self):
        """
        get_on_nodes gets the id of all nodes that haven't been seen offline in the last
        completed week. They were seen online the whole time.
        """
        cur = self.conn.cursor()
        cur.execute(
            """
            SELECT count(DISTINCT peer_id)
            FROM sessions
            WHERE created_at < date_trunc('week', NOW())
              AND updated_at > date_trunc('week', NOW() - '1 week'::interval)
              AND (first_failed_dial > date_trunc('week', NOW()) OR finished = false)
            """
        )
        return [i for sub in cur.fetchall() for i in sub]

from lib_db import DBClient, cache


class DBClientFilecoin(DBClient):
    complex_regex = "^(.*)-(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?(?:\+(.*))$"
    simple_regex = "^(.*)-(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*).*"

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
            SELECT CONCAT(av.agent_name, '-', av.agent_version), count(DISTINCT peer_id) "count"
            FROM visits v
                     INNER JOIN (
                SELECT id,
                       COALESCE(complex_regex[1], simple_regex[1]) agent_name,
                       CASE
                           WHEN complex_regex[1] IS NOT NULL THEN
                               CONCAT(complex_regex[2], '.', complex_regex[3], '.', complex_regex[4])
                           ELSE
                               CONCAT(simple_regex[2], '.', simple_regex[3], '.', simple_regex[4])
                           END                                     agent_version
                FROM (
                         SELECT id,
                                agent_version,
                                regexp_matches(agent_version, '{self.complex_regex}', 'i') complex_regex,
                                regexp_matches(agent_version, '{self.simple_regex}', 'i') simple_regex
                         FROM agent_versions)
                         AS semver
            ) av on av.id = v.agent_version_id
            WHERE v.created_at > {self.start}
              AND v.created_at < {self.end}
              AND v.type = 'crawl'
              AND v.error IS NULL
            GROUP BY CONCAT(av.agent_name, '-', av.agent_version)
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
            SELECT CONCAT(av.agent_name, '-', av.agent_version), count(DISTINCT peer_id) "count"
            FROM visits v
                     INNER JOIN (
                SELECT id,
                       COALESCE(complex_regex[1], simple_regex[1]) agent_name,
                       CASE
                           WHEN complex_regex[1] IS NOT NULL THEN
                               CONCAT(complex_regex[2], '.', complex_regex[3], '.', complex_regex[4])
                           ELSE
                               CONCAT(simple_regex[2], '.', simple_regex[3], '.', simple_regex[4])
                           END                                     agent_version
                FROM (
                         SELECT id,
                                agent_version,
                                regexp_matches(agent_version, '{self.complex_regex}', 'i') complex_regex,
                                regexp_matches(agent_version, '{self.simple_regex}', 'i') simple_regex
                         FROM agent_versions)
                         AS semver
            ) av on av.id = v.agent_version_id
            WHERE v.created_at > {self.start}
              AND v.created_at < {self.end}
              AND v.type = 'crawl'
              AND v.error IS NULL
              AND v.peer_id IN ({self.fmt_list(peer_ids)})
            GROUP BY CONCAT(av.agent_name, '-', av.agent_version)
            ORDER BY count DESC
            """
        )
        return cur.fetchall()

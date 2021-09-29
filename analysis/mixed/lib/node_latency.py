# get_latency gets the average latencies.
def get_latencies(conn):
    cur = conn.cursor()
    res = dict()
    cur.execute(
        """
        SELECT ping_latency_s_avg
        FROM latencies
        WHERE ping_latency_s_avg > 0 and updated_at > NOW() - '1 day'::interval
        """
    )
    return cur.fetchall()

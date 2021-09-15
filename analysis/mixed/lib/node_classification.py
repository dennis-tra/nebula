import pytz
import datetime
import psycopg2

def get_all_nodes(conn, start, end):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM sessions
        WHERE created_at < %s AND updated_at > %s
        """,
        [end, start]
    )
    return [i for sub in cur.fetchall() for i in sub]

def get_on_nodes(conn, start, end):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM sessions
        WHERE created_at < %s AND updated_at > %s AND first_successful_dial != last_successful_dial AND peer_id NOT IN (
            SELECT peer_id
            FROM sessions
            WHERE updated_at > %s AND updated_At < %s
        )
        """,
        [end, end, start, end]
    )
    return [i for sub in cur.fetchall() for i in sub]

def get_off_nodes(conn, start, end):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM sessions
        WHERE created_at < %s AND updated_at > %s AND first_successful_dial = last_successful_dial AND finished = true AND peer_id NOT IN (
            SELECT peer_id
            FROM sessions
            WHERE updated_at > %s AND updated_At < %s AND first_successful_dial != last_successful_dial
        )
        """,
        [end, start, start, end]
    )
    return [i for sub in cur.fetchall() for i in sub]

def get_dangling_nodes(conn, start, end):
    start = start.astimezone(pytz.utc)
    end = end.astimezone(pytz.utc)
    cur = conn.cursor()
    cur.execute(
        """
        SELECT DISTINCT peer_id
        FROM sessions
        WHERE updated_at > %s AND updated_at < %s AND first_successful_dial != last_successful_dial
        """,
        [start, end]
    )
    return [i for sub in cur.fetchall() for i in sub]

conn = psycopg2.connect(
    host="localhost",
    port=5432,
    database="nebula",
    user="nebula",
    password="password",
)
start = datetime.datetime.strptime("2021-09-15 18:00:00", "%Y-%m-%d %H:%M:%S").astimezone()
start = start.astimezone(pytz.utc)
end = datetime.datetime.strptime("2021-09-16 04:00:00", "%Y-%m-%d %H:%M:%S").astimezone()
end = end.astimezone(pytz.utc)

all = set(get_all_nodes(conn, start, end))
print(len(all))

off = set(get_off_nodes(conn, start, end))
print(len(off))

on = set(get_on_nodes(conn, start, end))
print(len(on))


dangle = set(get_dangling_nodes(conn, start, end))
print(len(dangle))

dangleC = all.difference(set.union(off, on))
print(len(dangleC))

diff = dangleC.difference(dangle)
print(diff)

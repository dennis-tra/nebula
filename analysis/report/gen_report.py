import jinja2
import datetime
from datetime import datetime as dt

from lib_fmt import fmt_thousands
from lib_db import DBClient

year = 2021
calendar_week = (datetime.date.today() - datetime.timedelta(weeks=1)).isocalendar().week

client = DBClient()
crawl_count = client.query(
    f"""
    SELECT count(*)
    FROM crawls c
    WHERE created_at > {client.start}
      AND created_at < {client.end}
    """
)

visit_count = client.query(
    f"""
    SELECT count(*)
    FROM visits v
    WHERE created_at > {client.start}
      AND created_at < {client.end}
    """
)

peer_id_count = client.query(
    f"""
    SELECT count(DISTINCT peer_id)
    FROM visits v
    WHERE created_at > {client.start}
      AND created_at < {client.end}
    """
)

new_agent_versions = client.query(
    f"""
    SELECT EXTRACT('epoch' FROM av.created_at), av.agent_version
    FROM agent_versions av
    WHERE created_at > {client.start}
      AND created_at < {client.end}
    ORDER BY av.created_at
    """
)

new_agent_versions = [f"`{av[1]}` ({dt.utcfromtimestamp(av[0]).strftime('%Y-%m-%d %H:%M:%S')})" for av in
                      new_agent_versions]

new_protocols = client.query(
    f"""
    SELECT EXTRACT('epoch' FROM p.created_at), p.protocol
    FROM protocols p
    WHERE created_at > {client.start}
      AND created_at < {client.end}
    ORDER BY p.created_at
    """
)
new_protocols = [f"`{p[1]}` ({dt.utcfromtimestamp(p[0]).strftime('%Y-%m-%d %H:%M:%S')})" for p in new_protocols]

loader = jinja2.FileSystemLoader(searchpath="./")
env = jinja2.Environment(loader=loader)
template = env.get_template("REPORT.tpl.md")
outputText = template.render(
    year=year,
    calendar_week=calendar_week,
    measurement_start=datetime.datetime.strptime(f"{year}-W{calendar_week}" + '-1', "%Y-W%W-%w").date(),
    measurement_end=datetime.datetime.strptime(f"{year}-W{calendar_week + 1}" + '-1', "%Y-W%W-%w").date(),
    crawl_count=fmt_thousands(crawl_count[0][0]),
    visit_count=fmt_thousands(visit_count[0][0]),
    peer_id_count=fmt_thousands(peer_id_count[0][0]),
    new_agent_versions=new_agent_versions,
    new_protocols=new_protocols
)

with open(f"report-{calendar_week}.md", "w") as f:
    f.write(outputText)

import jinja2
import datetime
from datetime import datetime as dt

from lib_cloud import Cloud
from lib_fmt import fmt_thousands
from lib_db import DBClient

year = 2021
calendar_week = (datetime.date.today() - datetime.timedelta(weeks=1)).isocalendar().week

db_client = DBClient()
cloud_client = Cloud()

crawl_count = db_client.query(
    f"""
    SELECT count(*)
    FROM crawls c
    WHERE created_at > {db_client.start}
      AND created_at < {db_client.end}
    """
)

visit_count = db_client.query(
    f"""
    SELECT count(*)
    FROM visits v
    WHERE created_at > {db_client.start}
      AND created_at < {db_client.end}
    """
)

peer_id_count = db_client.query(
    f"""
    SELECT count(DISTINCT peer_id)
    FROM visits v
    WHERE created_at > {db_client.start}
      AND created_at < {db_client.end}
    """
)

ip_address_count = db_client.query(
    f"""
    WITH peer_maddrs AS (
        SELECT v.peer_id, unnest(mas.multi_address_ids) multi_address_id
        FROM visits v
                 INNER JOIN multi_addresses_sets mas on mas.id = v.multi_addresses_set_id
                 LEFT OUTER JOIN agent_versions av on av.id = v.agent_version_id
        WHERE v.created_at > {db_client.start}
          AND v.created_at < {db_client.end}
        GROUP BY v.peer_id, unnest(mas.multi_address_ids)
    )
    SELECT count(DISTINCT ia.address)
    FROM multi_addresses ma
             INNER JOIN peer_maddrs pm ON pm.multi_address_id = ma.id
             INNER JOIN multi_addresses_x_ip_addresses maxia on pm.multi_address_id = maxia.multi_address_id
             INNER JOIN ip_addresses ia on maxia.ip_address_id = ia.id
    """
)

new_protocols = db_client.query(
    f"""
    SELECT EXTRACT('epoch' FROM p.created_at), p.protocol
    FROM protocols p
    WHERE created_at > {db_client.start}
      AND created_at < {db_client.end}
    ORDER BY p.created_at
    """
)
new_protocols = [f"`{p[1]}` ({dt.utcfromtimestamp(p[0]).strftime('%Y-%m-%d %H:%M:%S')})" for p in new_protocols]

from plot_agent_filecoin import main as plot_agent
from plot_cdf_arrivaltime_dangle_filecoin import main as plot_cdf_arrivaltime_dangle
from plot_churn_filecoin import main as plot_churn
from plot_cloud import main as plot_cloud
from plot_cloud_agents import main as plot_cloud_agents
from plot_cloud_classification import main as plot_cloud_classification
from plot_crawl import main as plot_crawl
from plot_crawl_properties import main as plot_crawl_properties
from plot_geo_agents_filecoin import main as plot_geo_agents
from plot_geo_classification import main as plot_geo_classification
from plot_geo_unique_ip import main as plot_geo_unique_ip
from plot_latencies import main as plot_latencies
from plot_latencies_geo import main as plot_latencies_geo
from plot_node_classification import main as plot_nodes

print("Running plot_agent...")
plot_agent(db_client)
print("Running plot_cdf_arrivaltime_dangle...")
plot_cdf_arrivaltime_dangle(db_client)
print("Running plot_churn...")
plot_churn(db_client)
print("Running plot_cloud...")
plot_cloud(db_client, cloud_client)
print("Running plot_cloud_agents...")
plot_cloud_agents(db_client, cloud_client)
print("Running plot_cloud_classification...")
plot_cloud_classification(db_client, cloud_client)
print("Running plot_crawl...")
plot_crawl(db_client)
print("Running plot_crawl_properties...")
plot_crawl_properties(db_client)
print("Running plot_geo_agents...")
plot_geo_agents(db_client)
print("Running plot_geo_classification...")
plot_geo_classification(db_client)
print("Running plot_geo_unique_ip...")
plot_geo_unique_ip(db_client, 20)
print("Running plot_latencies...")
plot_latencies(db_client)
print("Running plot_latencies_geo...")
plot_latencies_geo(db_client)
print("Running plot_nodes...")
plot_nodes(db_client)

loader = jinja2.FileSystemLoader(searchpath="./")
env = jinja2.Environment(loader=loader)
template = env.get_template("REPORT_filecoin.tpl.md")
outputText = template.render(
    year=year,
    calendar_week=calendar_week,
    measurement_start=datetime.datetime.strptime(f"{year}-W{calendar_week}" + '-1', "%Y-W%W-%w").date(),
    measurement_end=datetime.datetime.strptime(f"{year}-W{calendar_week + 1}" + '-1', "%Y-W%W-%w").date(),
    crawl_count=fmt_thousands(crawl_count[0][0]),
    visit_count=fmt_thousands(visit_count[0][0]),
    peer_id_count=fmt_thousands(peer_id_count[0][0]),
    new_protocols=new_protocols,
    ip_address_count=fmt_thousands(ip_address_count[0][0]),
)

with open(f"reports/report-{calendar_week}.md", "w") as f:
    f.write(outputText)

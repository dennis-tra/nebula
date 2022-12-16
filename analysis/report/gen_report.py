import jinja2
import datetime
from datetime import datetime as dt

from lib.lib_cloud import Cloud
from lib.lib_fmt import fmt_thousands
from lib.lib_db import DBClient

year = 2021
calendar_week = (datetime.date.today() - datetime.timedelta(weeks=1)).isocalendar().week

db_client = DBClient()
cloud_client = Cloud()

from plots.plot_cdf_arrivaltime_dangle import main as plot_cdf_arrivaltime_dangle
from plots.plot_churn import main as plot_churn
from plots.plot_cloud import main as plot_cloud
from plots.plot_cloud_agents import main as plot_cloud_agents
from plots.plot_cloud_classification import main as plot_cloud_classification
from plots.plot_crawl_overview import main as plot_crawl
from plots.plot_crawl_properties import main as plot_crawl_properties
from plots.plot_geo_agents import main as plot_geo_agents
from plots.plot_geo_classification import main as plot_geo_classification
from plots.plot_geo_resolution import main as plot_geo_resolution
from plots.plot_geo_unique_ip import main as plot_geo_unique_ip
from plots.plot_latencies import main as plot_latencies
from plots.plot_latencies_geo import main as plot_latencies_geo
from plots.plot_node_classifications import main as plot_nodes

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
print("Running plot_crawl_properties...")
plot_crawl_properties(db_client)
print("Running plot_geo_agents...")
plot_geo_agents(db_client)
print("Running plot_geo_classification...")
plot_geo_classification(db_client)
print("Running plot_geo_resolution...")
plot_geo_resolution(db_client)
print("Running plot_geo_unique_ip...")
plot_geo_unique_ip(db_client)
print("Running plot_latencies...")
plot_latencies(db_client)
print("Running plot_latencies_geo...")
plot_latencies_geo(db_client)
print("Running plot_nodes...")
plot_nodes(db_client)

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
    new_protocols=new_protocols,
    top_rotating_hosts=top_rotating_hosts,
    ip_address_count=fmt_thousands(ip_address_count[0][0]),
    top_updating_hosts=top_updating_hosts,
)

with open(f"report-{calendar_week}.md", "w") as f:
    f.write(outputText)

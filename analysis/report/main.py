import seaborn as sns
from datetime import datetime as dt
from pandas.io.formats.style import jinja2

from lib import DBClient, lib_plot, NodeClassification
from lib.lib_fmt import fmt_thousands, fmt_discovered_entity

from plots import *


def generate_ipfs_report():
    sns.set_theme()

    now = dt.today()
    # TODO: subtract one week

    year = now.year
    calendar_week = now.isocalendar().week
    db_client = DBClient(year=year, calendar_week=calendar_week)

    ##################################
    crawl_count = db_client.get_crawl_count()
    visit_count = db_client.get_visit_count()
    peer_id_count = db_client.get_peer_id_count()
    ip_address_count = db_client.get_ip_addresses_count()

    new_agent_versions = db_client.get_new_agent_versions()
    new_agent_versions_strs = fmt_discovered_entity(new_agent_versions)

    new_protocols = db_client.get_new_protocols()
    new_protocols_strs = fmt_discovered_entity(new_protocols)

    top_rotating_hosts = db_client.get_top_rotating_hosts()
    top_updating_hosts = db_client.get_top_updating_hosts()

    ##################################
    df = db_client.get_agent_versions_distribution()
    fig = plot_agents_overall(df)
    lib_plot.savefig(fig, "agents-overall", db_client.calendar_week)

    ##################################
    fig = plot_agents_kubo(df)
    lib_plot.savefig(fig, "agents-kubo", db_client.calendar_week)

    ##################################
    node_classes = [
        NodeClassification.DANGLING,
        NodeClassification.ONLINE,
        NodeClassification.ONEOFF,
        NodeClassification.ENTERED
    ]

    for node_class in node_classes:
        peer_ids = db_client.node_classification_funcs[node_class]()
        if len(peer_ids) == 0:
            print(f"skipping {str(node_class.name).lower()} agent versions plot")
            continue
        df = db_client.get_agent_versions_for_peer_ids(peer_ids)
        fig = plot_agents_overall(df)
        lib_plot.savefig(fig, f"agents-{str(node_class.name).lower()}", db_client.calendar_week)

    ##################################
    all_peer_ids, data = data_node_classifications(db_client)
    fig = plot_node_classifications(all_peer_ids, data)
    lib_plot.savefig(fig, "node-classifications", db_client.calendar_week)

    ##################################
    fig = plot_crawl_overview(db_client.get_crawls())
    lib_plot.savefig(fig, "crawl-overview", db_client.calendar_week)

    ##################################
    fig = plot_crawl_properties(db_client.get_crawl_properties())
    lib_plot.savefig(fig, "crawl-properties", db_client.calendar_week)

    db_client.close()

    loader = jinja2.FileSystemLoader(searchpath="./")
    env = jinja2.Environment(loader=loader)
    template = env.get_template("REPORT.tpl.md")
    outputText = template.render(
        year=year,
        calendar_week=calendar_week,
        measurement_start=dt.strptime(f"{year}-W{calendar_week}" + '-1', "%Y-W%W-%w").date(),
        measurement_end=dt.strptime(f"{year}-W{calendar_week + 1}" + '-1', "%Y-W%W-%w").date(),
        crawl_count=fmt_thousands(crawl_count),
        visit_count=fmt_thousands(visit_count),
        peer_id_count=fmt_thousands(peer_id_count),
        new_agent_versions=new_agent_versions_strs,
        new_protocols=new_protocols_strs,
        top_rotating_hosts=top_rotating_hosts,
        ip_address_count=fmt_thousands(ip_address_count),
        top_updating_hosts=top_updating_hosts,
    )

    with open(f"report-{calendar_week}.md", "w") as f:
        f.write(outputText)


if __name__ == '__main__':
    generate_ipfs_report()

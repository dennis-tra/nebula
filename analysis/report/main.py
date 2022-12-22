import seaborn as sns
import datetime as dt
from pandas.io.formats.style import jinja2

from lib import DBClient, lib_plot, NodeClassification
from lib.lib_fmt import fmt_thousands

from plots import *


def generate_ipfs_report():
    sns.set_theme()

    now = dt.datetime.today()

    year = now.year
    calendar_week = now.isocalendar().week - 1
    db_client = DBClient(year=year, calendar_week=calendar_week)

    ### TEMPORARY START
    db_client.start_date = dt.datetime.strptime(f"2022-12-14 12:57:30", "%Y-%m-%d %H:%M:%S")
    db_client.end_date = dt.datetime.strptime(f"2022-12-19 00:00:00", "%Y-%m-%d %H:%M:%S")

    db_client.start = f"'{db_client.start_date.strftime('%Y-%m-%d %H:%M:%S')}'::TIMESTAMP"
    db_client.end = f"'{db_client.end_date.strftime('%Y-%m-%d %H:%M:%S')}'::TIMESTAMP"
    db_client.range = f"'[{db_client.start_date.strftime('%Y-%m-%d %H:%M:%S')}, {db_client.end_date.strftime('%Y-%m-%d %H:%M:%S')})'::TSTZRANGE"
    ### TEMPORARY END

    ##################################
    crawl_count = db_client.get_crawl_count()
    visit_count = db_client.get_visit_count()
    peer_id_count = db_client.get_peer_id_count()
    ip_address_count = db_client.get_ip_addresses_count()

    top_rotating_nodes = db_client.get_top_rotating_nodes()
    top_updating_nodes = db_client.get_top_updating_nodes()

    ##################################
    df = db_client.get_agent_versions_distribution()
    fig = plot_agents_kubo(df)
    lib_plot.savefig(fig, "agents-kubo", db_client.calendar_week)

    ##################################
    fig = plot_agents_overall(df)
    lib_plot.savefig(fig, "agents-overall", db_client.calendar_week)

    ##################################
    node_classes = [
        NodeClassification.DANGLING,
        NodeClassification.ONLINE,
        NodeClassification.ONEOFF,
        NodeClassification.ENTERED
    ]

    for node_class in node_classes:
        peer_ids = db_client.node_classification_funcs[node_class]()
        df = db_client.get_agent_versions_for_peer_ids(peer_ids)
        if len(df) == 0:
            continue
        fig = plot_agents_overall(df)
        lib_plot.savefig(fig, f"agents-{str(node_class.name).lower()}", db_client.calendar_week)

    ##################################
    all_peer_ids, data = data_node_classifications(db_client)
    fig = plot_peer_classifications(all_peer_ids, data)
    lib_plot.savefig(fig, "peer-classifications", db_client.calendar_week)

    ##################################
    fig = plot_crawl_overview(db_client.get_crawls())
    lib_plot.savefig(fig, "crawl-overview", db_client.calendar_week)

    ##################################
    fig = plot_crawl_properties(db_client.get_crawl_properties())
    lib_plot.savefig(fig, "crawl-properties", db_client.calendar_week)

    ##################################
    fig = plot_churn(db_client.get_peer_uptime(), int((db_client.half_date-db_client.start_date).seconds/60/60))
    lib_plot.savefig(fig, "peer-churn", db_client.calendar_week)

    ##################################
    data = db_client.get_inter_arrival_time(db_client.get_dangling_peer_ids())
    fig = plot_inter_arrival_time(data)
    lib_plot.savefig(fig, "peer-inter-arrival-time", calendar_week=calendar_week)

    ##################################
    data = db_client.get_geo_ip_addresses()
    fig = plot_geo_unique_ip(data)
    lib_plot.savefig(fig, "geo-unique-ip", calendar_week=calendar_week)

    ##################################
    countries = db_client.get_countries()
    country_distributions = {}
    for node_class in NodeClassification:
        peer_ids = db_client.node_classification_funcs[node_class]()
        data = countries[countries["peer_id"].isin(peer_ids)]
        data = data.groupby(by="country", as_index=False).count().sort_values('peer_id', ascending=False)
        data = data.rename(columns={'peer_id': 'count'})
        country_distributions[node_class] = data

    fig = plot_geo_classification(country_distributions)
    lib_plot.savefig(fig, "geo-peer-classification", calendar_week=calendar_week)

    ##################################
    peer_id_agents = db_client.get_peer_id_agent_versions()
    fig = plot_geo_agents(peer_id_agents, countries)
    lib_plot.savefig(fig, "geo-peer-agents", calendar_week=calendar_week)


    loader = jinja2.FileSystemLoader(searchpath="./")
    env = jinja2.Environment(loader=loader)
    template = env.get_template("REPORT.tpl.md")
    outputText = template.render(
        year=year,
        calendar_week=calendar_week,
        measurement_start=dt.datetime.strptime(f"{year}-W{calendar_week}" + '-1', "%Y-W%W-%w").date(),
        measurement_end=dt.datetime.strptime(f"{year}-W{calendar_week + 1}" + '-1', "%Y-W%W-%w").date(),
        crawl_count=fmt_thousands(crawl_count),
        visit_count=fmt_thousands(visit_count),
        peer_id_count=fmt_thousands(peer_id_count),
        storm_agent_versions=db_client.get_storm_agent_versions(),
        new_agent_versions=db_client.get_new_agent_versions(),
        new_protocols=db_client.get_new_protocols(),
        top_rotating_nodes=top_rotating_nodes,
        ip_address_count=fmt_thousands(ip_address_count),
        top_updating_nodes=top_updating_nodes,
    )

    with open(f"report-{calendar_week}.md", "w") as f:
        f.write(outputText)

    db_client.close()


if __name__ == '__main__':
    generate_ipfs_report()

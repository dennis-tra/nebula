from typing import Dict

import os
import sys
import pandas as pd
import seaborn as sns
import datetime as dt
from pandas.io.formats.style import jinja2

from lib import DBClient, lib_plot, NodeClassification
from lib.lib_fmt import fmt_thousands
from lib.lib_udger import UdgerClient

from plots import *


def generate_ipfs_report():
    sns.set_theme()

    now = dt.datetime.today()
    year = os.getenv('NEBULA_REPORT_YEAR', now.year)
    calendar_week = os.getenv('NEBULA_REPORT_WEEK', now.isocalendar().week - 1)

    if len(sys.argv) > 1:
        output_dir = sys.argv[1]
    else:
        output_dir = '.'

    # allow name of plots directory to be overridden
    plots_dirname = os.getenv('NEBULA_PLOTS_DIRNAME', f"plots-{calendar_week}")
    output_file = os.path.join(output_dir, f"report-{calendar_week}.md")
    plots_dir = os.path.join(output_dir, plots_dirname)
    if not os.path.isdir(plots_dir):
        os.mkdir(plots_dir)

    db_config = {
        'host': os.environ['NEBULA_DATABASE_HOST'],
        'port': os.environ['NEBULA_DATABASE_PORT'],
        'database': os.environ['NEBULA_DATABASE_NAME'],
        'user': os.environ['NEBULA_DATABASE_USER'],
        'password': os.environ['NEBULA_DATABASE_PASSWORD'],
        'sslmode': os.environ['NEBULA_DATABASE_SSL_MODE'],
    }

    udger_db_file=os.getenv('NEBULA_REPORT_UDGER_DB', 'udgerdb_v3.dat')
    udger_available=os.path.exists(udger_db_file)

    print(f'Generating report for year {year}, week {calendar_week}')
    print(f'Writing report to {output_file}')
    print(f'Writing plots to {plots_dir}')
    print('Using database connection with:')
    print('host: ' + db_config['host'])
    print('port: ' + db_config['port'])
    print('database: ' + db_config['database'])
    print('user: ' + db_config['user'])
    print('sslmode: ' + db_config['sslmode'])


    db_client = DBClient(year=year, calendar_week=calendar_week, db_config=db_config)
    if udger_available:
        print('Using udger database')
        udger_client = UdgerClient(udger_db_file)

    ##################################
    crawl_count = db_client.get_crawl_count()
    visit_count = db_client.get_visit_count()
    visited_peer_id_count = db_client.get_visited_peer_id_count()
    discovered_peer_id_count = db_client.get_discovered_peer_id_count()
    ip_address_count = db_client.get_ip_addresses_count()

    top_rotating_nodes = db_client.get_top_rotating_nodes()
    # top_updating_peers = db_client.get_top_updating_peers()

    ##################################
    fig = plot_crawl_errors(db_client.get_connection_errors(), db_client.get_crawl_errors())
    lib_plot.savefig(fig, "crawl-errors", dir_name=plots_dir)

    ##################################
    df = db_client.get_agent_versions_distribution()
    fig = plot_agents_kubo(df)
    lib_plot.savefig(fig, "agents-kubo", dir_name=plots_dir)

    ##################################
    fig = plot_agents_overall(df)
    lib_plot.savefig(fig, "agents-overall", dir_name=plots_dir)

    ##################################
    node_classes = [
        NodeClassification.DANGLING,
        NodeClassification.ONLINE,
        NodeClassification.ONEOFF,
        NodeClassification.ENTERED
    ]

    agents = {}
    for node_class in node_classes:
        peer_ids = db_client.node_classification_funcs[node_class]()
        df = db_client.get_agent_versions_for_peer_ids(peer_ids)
        if len(df) == 0:
            continue
        agents[node_class.name] = df
    fig = plot_agents_classification(agents)
    lib_plot.savefig(fig, f"agents-classification", dir_name=plots_dir)

    ##################################
    all_peer_ids, data = data_node_classifications(db_client)
    fig = plot_peer_classifications(all_peer_ids, data)
    lib_plot.savefig(fig, "peer-classifications", dir_name=plots_dir)

    ##################################
    fig = plot_crawl_overview(db_client.get_crawls())
    lib_plot.savefig(fig, "crawl-overview", dir_name=plots_dir)

    ##################################
    fig = plot_crawl_properties(db_client.get_crawl_agent_versions())
    lib_plot.savefig(fig, "crawl-properties", dir_name=plots_dir)

    ##################################
    fig = plot_crawl_protocols(db_client.get_crawl_protocols())
    lib_plot.savefig(fig, "crawl-protocols", dir_name=plots_dir)

    ##################################
    fig = plot_crawl_unresponsive(db_client.get_unresponsive_peers_over_time())
    lib_plot.savefig(fig, "crawl-unresponsive", dir_name=plots_dir)

    ##################################
    node_classes = [
        NodeClassification.ONLINE,
        NodeClassification.OFFLINE,
        NodeClassification.DANGLING,
        NodeClassification.ONEOFF,
        NodeClassification.ENTERED,
        NodeClassification.LEFT
    ]

    classifications_over_time: Dict[NodeClassification, pd.DataFrame] = {}
    for node_class in node_classes:
        df = db_client.get_classification_over_time(node_class)
        classifications_over_time[node_class] = df
    fig = plot_crawl_classifications(classifications_over_time)
    lib_plot.savefig(fig, "crawl-classifications", dir_name=plots_dir)

    ##################################
    fig = plot_churn(db_client.get_peer_uptime(), int((db_client.half_date-db_client.start_date).seconds/60/60))
    lib_plot.savefig(fig, "peer-churn", dir_name=plots_dir)

    ##################################
    data = db_client.get_inter_arrival_time(db_client.get_dangling_peer_ids())
    fig = plot_inter_arrival_time(data)
    lib_plot.savefig(fig, "peer-inter-arrival-time", dir_name=plots_dir)

    ##################################
    data = db_client.get_geo_ip_addresses()
    fig = plot_geo_unique_ip(data)
    lib_plot.savefig(fig, "geo-unique-ip", dir_name=plots_dir)

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
    lib_plot.savefig(fig, "geo-peer-classification", dir_name=plots_dir)

    ##################################
    peer_id_agents = db_client.get_peer_id_agent_versions()
    fig = plot_geo_agents(peer_id_agents, countries)
    lib_plot.savefig(fig, "geo-peer-agents", dir_name=plots_dir)

    ##################################
    if udger_available:
        data = db_client.get_overall_cloud_distribution()
        data["datacenter"] = data.apply(lambda row: udger_client.get_datacenter(row["datacenter_id"]).name if udger_client.get_datacenter(row["datacenter_id"]) is not None else "Non-Datacenter", axis=1)
        data = data.drop(columns=["datacenter_id"])
        fig = plot_cloud_overall(data)
        lib_plot.savefig(fig, "cloud-overall", dir_name=plots_dir)

        ##################################
        peer_id_clouds = db_client.get_peer_id_cloud_distribution()
        peer_id_agents = db_client.get_peer_id_agent_versions()
        peer_id_clouds["datacenter"] = peer_id_clouds.apply(lambda row: udger_client.get_datacenter(row["datacenter_id"]).name if udger_client.get_datacenter(row["datacenter_id"]) is not None else "Non-Datacenter", axis=1)
        peer_id_clouds = peer_id_clouds.drop(columns=["datacenter_id"])
        fig = plot_cloud_agents(peer_id_agents, peer_id_clouds)
        lib_plot.savefig(fig, "cloud-agents", dir_name=plots_dir)

        ##################################
        node_classes = [
            NodeClassification.DANGLING,
            NodeClassification.ONLINE,
            NodeClassification.ONEOFF,
            NodeClassification.ENTERED,
            NodeClassification.LEFT,
        ]
        clouds_distributions = {}
        for node_class in node_classes:
            peer_ids = db_client.node_classification_funcs[node_class]()
            data = peer_id_clouds[peer_id_clouds["peer_id"].isin(peer_ids)]
            data = data.groupby(by="datacenter", as_index=False).count().sort_values('peer_id', ascending=False).reset_index(drop=True)
            data = data.rename(columns={'peer_id': 'count'})
            clouds_distributions[node_class] = data

        fig = plot_cloud_classification(clouds_distributions)
        lib_plot.savefig(fig, "cloud-classification", dir_name=plots_dir)

    loader = jinja2.FileSystemLoader(searchpath="./")
    env = jinja2.Environment(loader=loader)
    template = env.get_template("REPORT.tpl.md")
    storm_agent_versions = db_client.get_storm_agent_versions()

    print("Rendering template...")
    try:
        outputText = template.render(
            year=year,
            calendar_week=calendar_week,
            measurement_start=dt.datetime.strptime(f"{year}-W{calendar_week}" + '-1', "%Y-W%W-%w").date(),
            measurement_end=dt.datetime.strptime(f"{year}-W{calendar_week + 1}" + '-1', "%Y-W%W-%w").date(),
            crawl_count=fmt_thousands(crawl_count),
            visit_count=fmt_thousands(visit_count),
            visited_peer_id_count=fmt_thousands(visited_peer_id_count),
            discovered_peer_id_count=fmt_thousands(discovered_peer_id_count),
            storm_agent_versions=storm_agent_versions,
            storm_star_agent_versions=[av for av in storm_agent_versions if av != "storm"],
            new_agent_versions=db_client.get_new_agent_versions(),
            new_protocols=db_client.get_new_protocols(),
            top_rotating_nodes=top_rotating_nodes,
            ip_address_count=fmt_thousands(ip_address_count),
            plots_dir=plots_dir,
            # top_updating_peers=top_updating_peers,
        )
    except Exception as e:
        print(e)


    print("Writing templated output...")
    try:
        with open(output_file, "w") as f:
            f.write(outputText)
    except Exception as e:
        print(e)


    db_client.close()


if __name__ == '__main__':
    generate_ipfs_report()

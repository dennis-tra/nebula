import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd

import lib_plot
from lib_db import DBClient
from lib_fmt import fmt_thousands, fmt_barplot


def main():
    sns.set_theme()

    client = DBClient()

    all_peer_ids = client.get_all_peer_ids()
    unresolved_peer_ids = client.get_unresolved_peer_ids()
    no_public_ip_peer_ids = client.get_no_public_ip_peer_ids()

    countries = pd.DataFrame(client.get_countries(), columns=["peer_id", "country"])
    resolved_peer_ids = set(countries["peer_id"].unique())

    countries_with_relays = pd.DataFrame(client.get_countries_with_relays(), columns=["peer_id", "country"])
    resolved_with_relays_peer_ids = set(countries_with_relays["peer_id"].unique())

    relay_only_peer_ids = resolved_with_relays_peer_ids - resolved_peer_ids

    data = pd.DataFrame.from_dict({
        'Classification': [
            "resolved",
            "unresolved",
            "no public ip",
            "relay only"
        ],
        'Count': [
            len(resolved_peer_ids),
            len(unresolved_peer_ids),
            len(no_public_ip_peer_ids),
            len(relay_only_peer_ids)
        ]
    })

    fig, ax = plt.subplots(figsize=(10, 5))

    sns.barplot(ax=ax, x="Classification", y="Count", data=data)
    fmt_barplot(ax, data["Count"], data["Count"].sum())

    assert data['Count'].sum() == len(all_peer_ids)

    plt.title(f"Peer ID to IP Address Resolution Classification (Total {fmt_thousands(data['Count'].sum())})")

    lib_plot.savefig(f"geo-new")
    plt.show()


if __name__ == '__main__':
    main()

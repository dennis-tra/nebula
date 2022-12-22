import pandas as pd
import numpy as np
import seaborn as sns
from matplotlib import pyplot as plt
import geoip2.database

import lib_plot
from lib_fmt import thousands_ticker_formatter, fmt_thousands
from lib_db import DBClient


def main(db_client: DBClient):
    sns.set_theme()

    dial_results = db_client.query(
        f"""
        WITH cte AS (
            SELECT v.id,
                   EXTRACT('epoch' FROM v.dial_duration) dial_duration,
                   unnest(mas.multi_address_ids)         multi_address_id
            FROM visits v
                     INNER JOIN multi_addresses_sets mas on v.multi_addresses_set_id = mas.id
            WHERE v.created_at > {db_client.start}
              AND v.created_at < {db_client.end}
              AND v.type = 'dial'
              AND v.error IS NULL
        )
        SELECT ia.address, cte.dial_duration
        FROM multi_addresses_x_ip_addresses maxia
                 INNER JOIN cte ON cte.multi_address_id = maxia.multi_address_id
                 INNER JOIN ip_addresses ia on maxia.ip_address_id = ia.id
        GROUP BY ia.address, cte.dial_duration
        """
    )

    dial_results_df = pd.DataFrame(dial_results, columns=["ip_address", "dial_duration"])

    with geoip2.database.Reader("./GeoLite2-Country.mmdb") as geoipreader:
        def determine_continent(address):
            try:
                return geoipreader.country(address).continent.code
            except:
                return None

        print(f"Assigning Continents {len(dial_results_df)}")
        dial_results_df = dial_results_df.assign(
            continent=lambda df: df.ip_address.apply(determine_continent),
        )

        continents = ["EU", "NA", "AS", "SA", "OC"]
        continents_mapping = {
            "EU": "Europe",
            "NA": "North America",
            "SA": "South America",
            "AS": "Asia",
            "OC": "Oceania"
        }

        print("Starting Plot")

        fig, axs = plt.subplots(len(continents), 1, figsize=(10, len(continents) * 3))

        for idx, ax in enumerate(axs):
            data = dial_results_df[dial_results_df["continent"] == continents[idx]]
            sns.histplot(ax=ax, data=data, x="dial_duration", bins=np.arange(0, 6, 0.05))

            ax.get_yaxis().set_major_formatter(thousands_ticker_formatter)
            ax.set_xlim(0)
            ax.set_xlabel("Time in s")
            ax.title.set_text(
                f"Dial Durations to {continents_mapping[continents[idx]]} (Total {fmt_thousands(len(data))})")

        plt.tight_layout()
        lib_plot.savefig("latencies-geo")
        plt.show()


if __name__ == '__main__':
    db_client = DBClient()
    main(db_client)

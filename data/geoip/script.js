const postgres = require("postgres");
const { Multiaddr } = require("multiaddr");
const fs = require("fs");
const Reader = require("@maxmind/geoip2-node").Reader;

const dbBuffer = fs.readFileSync("GeoLite2-Country.mmdb");
const reader = Reader.openBuffer(dbBuffer);

const sql = postgres("postgres://nebula:password@localhost:5432/nebula");

const countries = { unknown: [] };

const main = async () => {
    const peers = await sql`select * from peers`;
    for (const peer of peers) {
        let found = false;
        for (const maddrStr of peer.multi_addresses) {
            try {
                const maddr = new Multiaddr(maddrStr);
                console.log(maddr);
                const response = reader.country(maddr.nodeAddress().address);
                const iso = response.country.isoCode;
                if (countries[iso] === undefined) {
                    countries[iso] = 0;
                }
                countries[iso]++;
                found = true;
                break;
            } catch (error) {
                console.error(error);
            }
        }
        if (!found) {
            countries["unknown"].push(peer.multi_addresses);
        }
    }
    console.log(countries);
    return;
};

main();

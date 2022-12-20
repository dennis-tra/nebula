![Nebula Logo](./docs/nebula-logo.svg)

# Nebula

[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)
[![readme nebula](https://img.shields.io/badge/readme-Nebula-blueviolet)](README.md)
[![GitHub license](https://img.shields.io/github/license/dennis-tra/nebula)](https://github.com/dennis-tra/nebula/blob/main/LICENSE)
[![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fdennis-tra%2Fnebula&count_bg=%2379C83D&title_bg=%23555555&icon=&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://hits.seeyoufarm.com)

A libp2p DHT crawler that also monitors the liveness and availability of peers. The crawler connects to the standard [DHT](https://en.wikipedia.org/wiki/Distributed_hash_table) bootstrap nodes and then recursively follows all entries in their [k-buckets](https://en.wikipedia.org/wiki/Kademlia) until all peers have been visited. Currently I'm running it for the IPFS and Filecoin networks.

🏆 _The crawler was awarded a prize in the context of the [DI2F Workshop hackathon](https://research.protocol.ai/blog/2021/decentralising-the-internet-with-ipfs-and-filecoin-di2f-a-report-from-the-trenches/)._ 🏆

📊 A Demo Dashboard can be found [here](http://162.55.187.75:3000/d/CSQsORs7k/nebula)! 📊

![Screenshot from a Grafana dashboard](./docs/grafana-screenshot.png)

## Table of Contents

- [Project Status](#project-status)
- [Usage](#usage)
- [How does it work?](#how-does-it-work)
  - [`crawl`](#crawl) | [`monitor`](#monitor)| [`resolve`](#resolve)
- [Install](#install)
  - [Release download](#release-download) | [From source](#from-source)
- [Development](#development)
  - [Database](#database)
- [Deployment](#deployment)
- [Analysis](#analysis)
- [Related Efforts](#related-efforts)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [Support](#support)
- [Other Projects](#other-projects)
- [License](#license)

## Project Status

The crawler is used for a couple of academic project, and I'm running it since July '21 continuously.

The gathered numbers about the IPFS network are in line with existing data like from the [`wiberlin/ipfs-crawler`](https://github.com/wiberlin/ipfs-crawler). Their crawler also powers a dashboard which can be found [here](https://trudi.weizenbaum-institut.de/ipfs_analysis.html).

## Usage

Nebula is a command line tool and provides the `crawl` sub-command. To simply crawl the IPFS network run:

```shell
nebula crawl --dry-run
```

Usually the crawler will persist its result in a postgres database - the `--dry-run` flag prevents it from doing that. One run takes ~5-10 min depending on your internet connection. You can also specify the network you want to crawl by appending, e.g., `--network FILECOIN`.

See the command line help page below for global configuration options or consult the sub-command specific ones:

```shell
NAME:
   nebula - A libp2p DHT crawler, monitor, and measurement tool that exposes timely information about DHT networks.

USAGE:
   nebula [global options] command [command options] [arguments...]

VERSION:
   vdev+5f3759df

AUTHOR:
   Dennis Trautwein <nebula@dtrautwein.eu>

COMMANDS:
   crawl    Crawls the entire network starting with a set of bootstrap nodes.
   monitor  Monitors the network by periodically dialing previously crawled peers.
   resolve  Resolves all multi addresses to their IP addresses and geo location information
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug                                  Set this flag to enable debug logging (default: false) [$NEBULA_DEBUG]
   --log-level value                        Set this flag to a value from 0 (least verbose) to 6 (most verbose). Overrides the --debug flag (default: 4) [$NEBULA_LOG_LEVEL]
   --config FILE                            Load configuration from FILE [$NEBULA_CONFIG_FILE]
   --dial-timeout value                     How long should be waited before a dial is considered unsuccessful (default: 1m0s) [$NEBULA_DIAL_TIMEOUT]
   --prom-port value                        On which port should prometheus serve the metrics endpoint (default: 6666) [$NEBULA_PROMETHEUS_PORT]
   --prom-host value                        Where should prometheus serve the metrics endpoint (default: 0.0.0.0) [$NEBULA_PROMETHEUS_HOST]
   --pprof-port value                       Enable pprof profiling endpoint on given port (default: disabled) [$NEBULA_PPROF_PORT]
   --db-host value                          On which host address can nebula reach the database (default: 0.0.0.0) [$NEBULA_DATABASE_HOST]
   --db-port value                          On which port can nebula reach the database (default: 5432) [$NEBULA_DATABASE_PORT]
   --db-name value                          The name of the database to use (default: nebula) [$NEBULA_DATABASE_NAME]
   --db-password value                      The password for the database to use (default: password) [$NEBULA_DATABASE_PASSWORD]
   --db-user value                          The user with which to access the database to use (default: nebula) [$NEBULA_DATABASE_USER]
   --db-sslmode value                       The sslmode to use when connecting the the database (default: disable) [$NEBULA_DATABASE_SSL_MODE]
   --protocols value [ --protocols value ]  Comma separated list of protocols that this crawler should look for (default: IPFS DHT: ) [$NEBULA_PROTOCOLS]
   --agent-versions-cache-size value        The cache size to hold agent versions in memory (default: 200) [$NEBULA_AGENT_VERSIONS_CACHE_SIZE]
   --protocols-cache-size value             The cache size to hold protocols in memory (default: 100) [$NEBULA_PROTOCOLS_CACHE_SIZE]
   --protocols-set-cache-size value         The cache size to hold sets of protocols in memory (default: 200) [$NEBULA_PROTOCOLS_SET_CACHE_SIZE]
   --help, -h                               show help (default: false)
   --version, -v                            print the version (default: false)
```

## How does it work?

### `crawl`

The `crawl` sub-command starts by connecting to a set of bootstrap nodes and constructing the routing tables (kademlia _k_-buckets)
of the remote peers based on their [`PeerIds`](https://docs.libp2p.io/concepts/peer-id/). Then `nebula` builds
random `PeerIds` with a common prefix length (CPL) and asks each remote peer if they know any peers that are
closer to the ones `nebula` just constructed (XOR distance). This will effectively yield a list of all `PeerIds` that a peer has
in its routing table. The process repeats for all found peers until `nebula` does not find any new `PeerIds`.

This process is heavily inspired by the `basic-crawler` in [libp2p/go-libp2p-kad-dht](https://github.com/libp2p/go-libp2p-kad-dht/tree/master/crawler) from @aschmahmann.

Every peer that was visited is persisted in the database. The information includes latency measurements (dial/connect/crawl durations), current set of multi addresses, current agent version and current set of supported protocols. If the peer was dialable `nebula` will
also create a `session` instance that contains the following information:

```sql
CREATE TABLE sessions (
    -- A unique id that identifies this particular session
    id                      INT GENERATED ALWAYS AS IDENTITY,
    -- Reference to the remote peer ID.
    peer_id                 INT           NOT NULL,
    -- Timestamp of the first time we were able to visit that peer.
    first_successful_visit  TIMESTAMPTZ   NOT NULL,
    -- Timestamp of the last time we were able to visit that peer.
    last_successful_visit   TIMESTAMPTZ   NOT NULL,
    -- Timestamp when we should start visiting this peer again.
    next_visit_due_at       TIMESTAMPTZ,
    -- When did we notice that this peer is not reachable.
    first_failed_visit      TIMESTAMPTZ,
    -- When did we first notice that this peer is not reachable.
    last_failed_visit       TIMESTAMPTZ,
    -- When did we last visit this peer. For indexing purposes.
    last_visited_at         TIMESTAMPTZ   NOT NULL,
    -- When was this session instance updated the last time
    updated_at              TIMESTAMPTZ   NOT NULL,
    -- When was this session instance created
    created_at              TIMESTAMPTZ   NOT NULL,
    -- Number of successful visits in this session.
    successful_visits_count INTEGER       NOT NULL,
    -- The number of times this session went from pending to open again.
    recovered_count         INTEGER       NOT NULL,
    -- The state this session is in (open, pending, closed)
    -- open: currently considered online
    -- pending: peer missed a dial and is pending to be closed
    -- closed: peer is considered to be offline and session is complete
    state                   session_state NOT NULL,
    -- Number of failed visits before closing this session.
    failed_visits_count     SMALLINT      NOT NULL,
    -- What's the first error before we close this session.
    finish_reason           net_error,
    -- The uptime time range for this session measured from first- to last_successful_visit to
    uptime                  TSTZRANGE     NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_sessions_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (id, state, last_visited_at)

) PARTITION BY LIST (state);
```

> Columns are ordered to optimize for storage by minimizing padding.

At the end of each crawl `nebula` persists general statistics about the crawl like the total duration, dialable peers, encountered errors, agent versions etc...

> **Info:** You can use the `crawl` sub-command with the `--dry-run` option that skips any database operations.

Command line help page:

```text
NAME:
   nebula crawl - Crawls the entire network starting with a set of bootstrap nodes.

USAGE:
   nebula crawl [command options] [arguments...]

OPTIONS:
   --bootstrap-peers value [ --bootstrap-peers value ]  Comma separated list of multi addresses of bootstrap peers [$NEBULA_BOOTSTRAP_PEERS]
   --workers value                                      How many concurrent workers should dial and crawl peers. (default: 1000) [$NEBULA_CRAWL_WORKER_COUNT]
   --limit value                                        Only crawl the specified amount of peers (0 for unlimited) (default: 0) [$NEBULA_CRAWL_PEER_LIMIT]
   --dry-run                                            Don't persist anything to a database (you don't need a running DB) (default: false) [$NEBULA_CRAWL_DRY_RUN]
   --neighbors                                          Whether to persist all k-bucket entries of a particular peer at the end of a crawl. (default: false) [$NEBULA_CRAWL_NEIGHBORS]
   --network value                                      Which network should be crawled (IPFS, FILECOIN, KUSAMA, POLKADOT). Presets default bootstrap peers. (default: IPFS) [$NEBULA_CRAWL_NETWORK]
   --help, -h                                           show help (default: false)
```

### `monitor`

The `monitor` sub-command polls every 10 seconds all sessions from the database (see above) that are due to be dialed
in the next 10 seconds (based on the `NextDialAttempt` timestamp). It attempts to dial all peers using previously
saved multi-addresses and updates their `session` instances accordingly if they're dialable or not.

The `NextDialAttempt` timestamp is calculated based on the uptime that `nebula` has observed for that given peer.
If the peer is up for a long time `nebula` assumes that it stays up and thus decreases the dial frequency aka. sets
the `NextDialAttempt` timestamp to a time further in the future.

Command line help page:

```text
NAME:
   nebula monitor - Monitors the network by periodically dialing previously crawled peers.

USAGE:
   nebula monitor [command options] [arguments...]

OPTIONS:
   --workers value  How many concurrent workers should dial peers. (default: 1000) [$NEBULA_MONITOR_WORKER_COUNT]
   --help, -h       show help (default: false)
```

### `resolve`

The resolve sub-command goes through all multi addresses that are present in the database and resolves them to their respective IP-addresses. Behind one multi address can be multiple IP addresses due to, e.g., the [`dnsaddr` protocol](https://github.com/multiformats/multiaddr/blob/master/protocols/DNSADDR.md).
Further, it queries the GeoLite2 database from [Maxmind](https://www.maxmind.com/en/home) to extract country information about the IP addresses and [UdgerDB](https://udger.com/) to detect datacenters. The command saves all information alongside the resolved addresses.

Command line help page:

```text
NAME:
   nebula resolve - Resolves all multi addresses to their IP addresses and geo location information

USAGE:
   nebula resolve [command options] [arguments...]

OPTIONS:
   --udger-db value    Location of the Udger database v3 [$NEBULA_RESOLVE_UDGER_DB]
   --batch-size value  How many database entries should be fetched at each iteration (default: 100) [$NEBULA_RESOLVE_BATCH_SIZE]
   --help, -h          show help (default: false)
```

## Install

### From source

To compile it yourself run:

```shell
go install github.com/dennis-tra/nebula/cmd/nebula@latest # Go 1.19 or higher is required (may work with a lower version too)
```

Make sure the `$GOPATH/bin` is in your PATH variable to access the installed `nebula` executable.

## Development

To develop this project you need Go `> 1.16` and the following tools:

- [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) to manage the SQL migration `v4.15.2`
- [`volatiletech/sqlboiler`](https://github.com/volatiletech/sqlboiler) to generate Go ORM `v4.13.0`
- `docker` to run a local postgres instance

To install the necessary tools you can run `make tools`. This will use the `go install` command to download and install the tools into your `$GOPATH/bin` directory. So make sure you have it in your `$PATH` environment variable.

### Database

You need a running postgres instance to persist and/or read the crawl results. Run `make database` or use the following command to start a local instance of postgres:

```shell
docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=nebula -e POSTGRES_DB=nebula postgres:14
```

> **Info:** You can use the `crawl` sub-command with the `--dry-run` option that skips any database operations.

The default database settings are:

```
Name     = "nebula",
Password = "password",
User     = "nebula",
Host     = "localhost",
Port     = 5432,
```

To apply migrations then run:

```shell
# Up migrations
make migrate-up # runs: migrate -database 'postgres://nebula:password@localhost:5432/nebula?sslmode=disable' -path migrations up

# Down migrations
make migrate-down # runs: migrate -database 'postgres://nebula:password@localhost:5432/nebula?sslmode=disable' -path migrations down

# Generate the ORM with SQLBoiler
make models # runs: sqlboiler
# This will update all files in the `pkg/models` directory.
```
```shell
# Create new migration
migrate create -ext sql -dir migrations -seq some_migration_name
```

### Tests

To run the tests you need a running test database instance:

```shell
make database-test
go test ./...
```

## Deployment

First, you need to build the nebula docker image:

```shell
make docker
# OR
docker build . -t dennis-tra/nebula:latest
```

The `deploy` subfolder contains a `docker-compose` setup to get up and running quickly. It will start and configure `nebula` (monitoring mode), `postgres`, `prometheus` and `grafana`. The configuration can serve as a starting point to see how things fit together. Then you can start the aforementioned services by changing in the `./deploy` directory and running:

```shell
docker compose up 
```

A few seconds later you should be able to access Grafana at `localhost:3000`. The initial credentials are

```text
USERNAME: admin
PASSWORD: admin
```

There is one preconfigured dashboard in the `General` folder with the name `IPFS Dashboard`. To start a crawl that puts its results in the `docker compose` provisioned postgres database run:

```shell
./deploy/crawl.sh
# OR
docker run \
  --network nebula \
  --name nebula \
  --hostname nebula \
  dennis-tra/nebula:latest \
  nebula --db-host=postgres crawl
```

Currently, I'm running the crawler docker-less on a tiny VPS in a 30m interval. The corresponding crontab configuration is:

```text
*/30 * * * * /some/path/nebula crawl 2> /var/log/nebula/crawl-$(date "+\%w-\%H-\%M")-stderr.log 1> /var/log/nebula/crawl-$(date "+\%w-\%H-\%M")-stdout.log
```

The logs will rotate every 7 days.

--- 

To run the crawler for multiple DHTs the idea is to also start multiple instances of `nebula` with the corresponding configuration. For instance, I'm running the crawler for the IPFS and Filecoin networks. The monitoring commands look like this:

```shell
nebula --prom-port=6667 monitor # for IPFS
nebula --prom-port=6669 --config filecoin.json monitor # for Filecoin
```

The `filecoin.json` file contains the following content:

```json5
 {
  "BootstrapPeers": [
    "/ip4/3.224.142.21/tcp/1347/p2p/12D3KooWCVe8MmsEMes2FzgTpt9fXtmCY7wrq91GRiaC8PHSCCBj",
    "/ip4/107.23.112.60/tcp/1347/p2p/12D3KooWCwevHg1yLCvktf2nvLu7L9894mcrJR4MsBCcm4syShVc",
    "/ip4/100.25.69.197/tcp/1347/p2p/12D3KooWEWVwHGn2yR36gKLozmb4YjDJGerotAPGxmdWZx2nxMC4",
    "/ip4/3.123.163.135/tcp/1347/p2p/12D3KooWKhgq8c7NQ9iGjbyK7v7phXvG6492HQfiDaGHLHLQjk7R",
    "/ip4/18.198.196.213/tcp/1347/p2p/12D3KooWL6PsFNPhYftrJzGgF5U18hFoaVhfGk7xwzD8yVrHJ3Uc",
    "/ip4/18.195.111.146/tcp/1347/p2p/12D3KooWLFynvDQiUpXoHroV1YxKHhPJgysQGH2k3ZGwtWzR4dFH",
    "/ip4/52.77.116.139/tcp/1347/p2p/12D3KooWP5MwCiqdMETF9ub1P3MbCvQCcfconnYHbWg6sUJcDRQQ",
    "/ip4/18.136.2.101/tcp/1347/p2p/12D3KooWRs3aY1p3juFjPy8gPN95PEQChm2QKGUCAdcDCC4EBMKf",
    "/ip4/13.250.155.222/tcp/1347/p2p/12D3KooWScFR7385LTyR4zU1bYdzSiiAb5rnNABfVahPvVSzyTkR",
    "/ip4/47.115.22.33/tcp/41778/p2p/12D3KooWGhufNmZHF3sv48aQeS13ng5XVJZ9E6qy2Ms4VzqeUsHk",
    "/ip4/61.147.123.111/tcp/12757/p2p/12D3KooWGHpBMeZbestVEWkfdnC9u7p6uFHXL1n7m1ZBqsEmiUzz",
    "/ip4/61.147.123.121/tcp/12757/p2p/12D3KooWQZrGH1PxSNZPum99M1zNvjNFM33d1AAu5DcvdHptuU7u",
    "/ip4/3.129.112.217/tcp/1235/p2p/12D3KooWBF8cpp65hp2u9LK5mh19x67ftAam84z9LsfaquTDSBpt",
    "/ip4/36.103.232.198/tcp/34721/p2p/12D3KooWQnwEGNqcM2nAcPtRR9rAX8Hrg4k9kJLCHoTR5chJfz6d",
    "/ip4/36.103.232.198/tcp/34723/p2p/12D3KooWMKxMkD5DMpSWsW7dBddKxKT7L2GgbNuckz9otxvkvByP"
  ],
  "DialTimeout": 60000000000,
  "CrawlWorkerCount": 1000,
  "MonitorWorkerCount": 1000,
  "CrawlLimit": 0,
  "MinPingInterval": 30000000000,
  "PrometheusHost": "localhost",
  "PrometheusPort": 6668, // this is overwritten by the command line arg and only picked up by the crawl command
  "DatabaseHost": "localhost",
  "DatabasePort": 5432,
  "DatabaseName": "nebula_filecoin",
  "DatabasePassword": "<password>",
  "DatabaseUser": "nebula_filecoin",
  "Protocols": [
    "/fil/kad/testnetnet/kad/1.0.0"
  ]
}
```

This configuration is created upon the first start of nebula in `$XDG_CONFIG_HOME/nebula/config.json` and can be adapted from there. 

The corresponding crawl commands look like:

```shell
nebula crawl # for IPFS (uses defaults like prom-port 6666)
nebula --config filecoin.json crawl --network FILECOIN # for Filecoin (uses configuration prom-port 6668)
```

The `workers` flag configures the amount of concurrent connection/dials. I increased it until I didn't notice any performance improvement anymore.

So this is the prometheus ports configuration:

- `nebula crawl` - `ipfs` - `6666`
- `nebula monitor` - `ipfs` - `6667`
- `nebula crawl` - `filecoin` - `6668`
- `nebula monitor` - `filecoin` - `6669`

Furthermore, `nebula` has a flag called `--pprof-port`. If this flag is set it also serves `pprof` at `localhost:pprof-port` for debugging.


## Analysis

There is a top-level `analysis` folder that contains various scripts to help understand the gathered data. More information can be found in the respective subfolders README file. The following evaluations can be found there

- [`geoip`](./analysis/geoip) - Uses a [Maxmind](https://maxmind.com) database to map IP addresses to country ISO codes and plots the results.
- [`churn`](./analysis/churn) - Uses a `sessions` database dump to construct a CDF of peer session lengths.
- [`mixed`](./analysis/mixed) - Multiple plotting scripts for various metrics of interest. See [wcgcyx/nebula-crawler](https://github.com/wcgcyx/nebula-crawler) for plots as I have just copied the scripts from there.
- [`report`](./analysis/report) - A semi-automated set of scripts to generate the reports for [dennis-tra/nebula-crawler-reports](https://github.com/dennis-tra/nebula-crawler-reports)
- More to come...

## Related Efforts

- [`wiberlin/ipfs-crawler`](https://github.com/wiberlin/ipfs-crawler) - A crawler for the IPFS network, code for their paper ([arXiv](https://arxiv.org/abs/2002.07747)).
- [`adlrocha/go-libp2p-crawler`](https://github.com/adlrocha/go-libp2p-crawler) - Simple tool to crawl libp2p networks resources
- [`libp2p/go-libp2p-kad-dht`](https://github.com/libp2p/go-libp2p-kad-dht/tree/master/crawler) - Basic crawler for the Kademlia DHT implementation on go-libp2p.

## Maintainers

[@dennis-tra](https://github.com/dennis-tra).

## Contributing

Feel free to dive in! [Open an issue](https://github.com/dennis-tra/nebula/issues/new) or submit PRs.

## Support

It would really make my day if you supported this project through [Buy Me A Coffee](https://www.buymeacoffee.com/dennistra).

## Other Projects

You may be interested in one of my other projects:

- [`pcp`](https://github.com/dennis-tra/pcp) - Command line peer-to-peer data transfer tool based on [libp2p](https://github.com/libp2p/go-libp2p).
- [`image-stego`](https://github.com/dennis-tra/image-stego) - A novel way to image manipulation detection. Steganography-based image integrity - Merkle tree nodes embedded into image chunks so that each chunk's integrity can be verified on its own.
- [`antares`](https://github.com/dennis-tra/antares) - A gateway and pinning service probing tool.

## License

[Apache License Version 2.0](LICENSE) © Dennis Trautwein

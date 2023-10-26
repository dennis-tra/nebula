![Nebula Logo](./docs/nebula-logo.svg)

# Nebula

[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)
[![readme nebula](https://img.shields.io/badge/readme-Nebula-blueviolet)](README.md)
[![GitHub license](https://img.shields.io/github/license/dennis-tra/nebula)](https://github.com/dennis-tra/nebula/blob/main/LICENSE)
[![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fdennis-tra%2Fnebula&count_bg=%2379C83D&title_bg=%23555555&icon=&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://hits.seeyoufarm.com)

A libp2p DHT crawler and monitor that tracks the liveness of peers. The crawler connects to [DHT](https://en.wikipedia.org/wiki/Distributed_hash_table) bootstrap peers and then recursively follows all entries in their [k-buckets](https://en.wikipedia.org/wiki/Kademlia) until all peers have been visited. The crawler supports the [IPFS](https://ipfs.network), [Filecoin](https://filecoin.io), [Polkadot](https://polkadot.network/), [Kusama](https://kusama.network/), [Rococo](https://substrate.io/developers/rococo-network/), [Westend](https://wiki.polkadot.network/docs/maintain-networks#westend-test-network) networks [and more](https://github.com/dennis-tra/nebula/blob/a33a5fd493caaeb07e92ecc73c32ee87ae9e374f/pkg/config/config.go#L11).

_The crawler was:_

- 🏆 _awarded a prize in the [DI2F Workshop hackathon](https://research.protocol.ai/blog/2021/decentralising-the-internet-with-ipfs-and-filecoin-di2f-a-report-from-the-trenches/)._ 🏆
- 🎓 _used for the ACM SigCOMM'22 paper [Design and Evaluation of IPFS: A Storage Layer for the Decentralized Web](https://research.protocol.ai/publications/design-and-evaluation-of-ipfs-a-storage-layer-for-the-decentralized-web/trautwein2022.pdf)_ 🎓


📊 We publish weekly reports based on the crawl results [here](https://github.com/protocol/network-measurements/tree/master/reports)! 📊

![Screenshot from a Grafana dashboard](./docs/grafana-screenshot.png)

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Project Status](#project-status)
- [Usage](#usage)
- [Install](#install)
  - [From source](#from-source)
- [How does it work?](#how-does-it-work)
  - [`crawl`](#crawl)
  - [`monitor`](#monitor)
  - [`resolve`](#resolve)
- [Development](#development)
  - [Database](#database)
  - [Tests](#tests)
- [Report](#report)
- [Related Efforts](#related-efforts)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [Support](#support)
- [Other Projects](#other-projects)
- [License](#license)

## Project Status

The crawler is used for a couple of academic project, and I'm running it since July '21 continuously.

The gathered numbers about the IPFS network are in line with existing data like from the [`wiberlin/ipfs-crawler`](https://github.com/wiberlin/ipfs-crawler). Their crawler also powers a dashboard which can be found [here](https://trudi.weizenbaum-institut.de/ipfs_analysis.html).


## Install

### From source

To compile it yourself run:

```shell
go install github.com/dennis-tra/nebula-crawler/cmd/nebula@latest # Go 1.19 or higher is required (may work with a lower version too)
```

Make sure the `$GOPATH/bin` is in your PATH variable to access the installed `nebula` executable.

## Usage

Nebula is a command line tool and provides the `crawl` sub-command. To simply crawl the IPFS network run:

```shell
nebula crawl --dry-run
```

The crawler can store its results as JSON documents or in a postgres database - the `--dry-run` flag prevents it from doing either. Nebula will print a summary of the crawl at the end instead. A crawl takes ~5-10 min depending on your internet connection. You can also specify the network you want to crawl by appending, e.g., `--network FILECOIN` and limit the number of peers to crawl by providing the `--limit` flag with the value of, e.g., `1000`. Example:

```shell
nebula crawl --dry-run --network FILECOIN --limit 1000
```

To store crawl results as JSON files provide the `--json-out` command line flag like so:

```shell
nebula crawl --json-out ./results/
```

After the crawl has finished, you will find the JSON files in the `./results/` subdirectory.

When providing only the `--json-out` command line flag you will see that the `*_neighbors.json` document is empty. This document would contain the full routing table information of each peer in the network which is quite a bit of data (~250MB as of April '23) and is therefore disabled by default. To populate the document you'll need to pass the `--neighbors` flag to the `crawl` subcommand.

```shell
nebula crawl --neighbors --json-out ./results/
```

The routing table information forms a graph and graph visualization tools often operate with [adjacency lists](https://en.wikipedia.org/wiki/Adjacency_list). To convert the `*_neighbors.json` document to an adjacency list, you can use [`jq`](https://stedolan.github.io/jq/) and the following command:

```shell
jq -r '.NeighborIDs[] as $neighbor | [.PeerID, $neighbor] | @csv' ./results/2023-04-16T14:32_neighbors.json > ./results/2023-04-16T14:32_neighbors.csv
```

There are a few more command line flags that are documented when you run `nebula crawl --help`.

When Nebula is configured to store its results in a postgres database (see below), then it also tracks session information of remote peers.

## How does it work?

### `crawl`

The `crawl` sub-command starts by connecting to a set of bootstrap nodes and constructing the routing tables (kademlia _k_-buckets)
of these peers based on their [`PeerIDs`](https://docs.libp2p.io/concepts/peer-id/). Then `nebula` builds
random `PeerIDs` with common prefix lengths (CPL) that fall each peers buckets, and asks each remote peer if they know any peers that are
closer (XOR distance) to the ones `nebula` just constructed. This will effectively yield a list of all `PeerIDs` that a peer has
in its routing table. The process repeats for all found peers until `nebula` does not find any new `PeerIDs`.

This process is heavily inspired by the `basic-crawler` in [libp2p/go-libp2p-kad-dht](https://github.com/libp2p/go-libp2p-kad-dht/tree/master/crawler) from [@aschmahmann](https://github.com/aschmahmann).

If Nebula is configured to store its results in a database, every peer that was visited is persisted in it. The visit information includes latency measurements (dial/connect/crawl durations), current set of multi addresses, current agent version and current set of supported protocols. If the peer was dialable `nebula` will
also create a `session` instance that contains the following information:

```sql
CREATE TABLE sessions (
    -- A unique id that identifies this particular session
    id                      INT GENERATED ALWAYS AS IDENTITY,
    -- Reference to the remote peer ID. (database internal ID)
    peer_id                 INT           NOT NULL,
    -- Timestamp of the first time we were able to visit that peer.
    first_successful_visit  TIMESTAMPTZ   NOT NULL,
    -- Timestamp of the last time we were able to visit that peer.
    last_successful_visit   TIMESTAMPTZ   NOT NULL,
    -- Timestamp when we should start visiting this peer again.
    next_visit_due_at       TIMESTAMPTZ,
    -- When did we notice that this peer is not reachable.
    first_failed_visit      TIMESTAMPTZ,
    -- When did we first notice that this peer is not reachable anymore.
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

At the end of each crawl `nebula` persists general statistics about the crawl like the total duration, dialable peers, encountered errors, agent versions etc...

> **Info:** You can use the `crawl` sub-command with the `--dry-run` option that skips any database operations.

Command line help page:

```text
NAME:
   nebula crawl - Crawls the entire network starting with a set of bootstrap nodes.

USAGE:
   nebula crawl [command options] [arguments...]

OPTIONS:
   --bootstrap-peers value [ --bootstrap-peers value ]  Comma separated list of multi addresses of bootstrap peers (default: default IPFS) [$NEBULA_CRAWL_BOOTSTRAP_PEERS, $NEBULA_BOOTSTRAP_PEERS]
   --protocols value [ --protocols value ]              Comma separated list of protocols that this crawler should look for [$NEBULA_CRAWL_PROTOCOLS, $NEBULA_PROTOCOLS]
   --workers value                                      How many concurrent workers should dial and crawl peers. (default: 1000) [$NEBULA_CRAWL_WORKER_COUNT]
   --limit value                                        Only crawl the specified amount of peers (0 for unlimited) (default: 0) [$NEBULA_CRAWL_PEER_LIMIT]
   --dry-run                                            Don't persist anything (default: false) [$NEBULA_CRAWL_DRY_RUN]
   --json-out DIR                                       If set, stores the crawl results as JSON documents at DIR (takes precedence over database settings). [$NEBULA_CRAWL_JSON_OUT]
   --neighbors                                          Whether to persist all k-bucket entries of a particular peer at the end of a crawl. (default: false) [$NEBULA_CRAWL_NEIGHBORS]
   --check-exposed                                      Whether to check if the Kubo API is exposed. Checking also includes crawling the API. (default: false) [$NEBULA_CRAWL_CHECK_EXPOSED]
   --network value                                      Which network should be crawled (IPFS, FILECOIN, KUSAMA, POLKADOT). Presets default bootstrap peers and protocol. (default: "IPFS") [$NEBULA_CRAWL_NETWORK]
   --help, -h                                           show help
```

### `monitor`

The `monitor` sub-command polls every 10 seconds all sessions from the database (see above) that are due to be dialed
in the next 10 seconds (based on the `next_visit_due_at` timestamp). It attempts to dial all peers using previously
saved multi-addresses and updates their `session` instances accordingly if they're dialable or not.

The `next_visit_due_at` timestamp is calculated based on the uptime that `nebula` has observed for that given peer.
If the peer is up for a long time `nebula` assumes that it stays up and thus decreases the dial frequency aka. sets
the `next_visit_due_at` timestamp to a time further in the future.

Command line help page:

```text
NAME:
   nebula monitor - Monitors the network by periodically dialing previously crawled peers.

USAGE:
   nebula monitor [command options] [arguments...]

OPTIONS:
   --workers value  How many concurrent workers should dial peers. (default: 1000) [$NEBULA_MONITOR_WORKER_COUNT]
   --help, -h       show help
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

## Development

To develop this project you need Go `> 1.16` and the following tools:

- [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) to manage the SQL migration `v4.15.2`
- [`volatiletech/sqlboiler`](https://github.com/volatiletech/sqlboiler) to generate Go ORM `v4.14.2`
- `docker` to run a local postgres instance

To install the necessary tools you can run `make tools`. This will use the `go install` command to download and install the tools into your `$GOPATH/bin` directory. So make sure you have it in your `$PATH` environment variable.

### Database

You need a running postgres instance to persist and/or read the crawl results. Run `make database` or use the following command to start a local instance of postgres:

```shell
docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=nebula -e POSTGRES_DB=nebula postgres:14
```

> **Info:** You can use the `crawl` sub-command with the `--dry-run` option that skips any database operations or store the results as JSON files with the `--json-out` flag.

The default database settings are:

```
Name     = "nebula",
Password = "password",
User     = "nebula",
Host     = "localhost",
Port     = 5432,
```

Migrations are applied automatically when `nebula` starts and successfully establishes a database connection.

To run them manually you can run:

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
migrate create -ext sql -dir pkg/db/migrations -seq some_migration_name
```

### Tests

To run the tests you need a running test database instance:

```shell
make database-test
go test ./...
```

## Report

There is a top-level `report` folder that contains a script to generate a comprehensive data report.

## Related Efforts

- [`wiberlin/ipfs-crawler`](https://github.com/wiberlin/ipfs-crawler) - A crawler for the IPFS network, code for their paper ([arXiv](https://arxiv.org/abs/2002.07747)).
- [`adlrocha/go-libp2p-crawler`](https://github.com/adlrocha/go-libp2p-crawler) - Simple tool to crawl libp2p networks resources
- [`libp2p/go-libp2p-kad-dht`](https://github.com/libp2p/go-libp2p-kad-dht/tree/master/crawler) - Basic crawler for the Kademlia DHT implementation on go-libp2p.
- [`migalabs/armiarma](https://github.com/migalabs/armiarma) - Armiarma is a Libp2p open-network crawler with a current focus on Ethereum's CL network
- [`migalabs/eth-light-crawler`](https://github.com/migalabs/eth-light-crawler) - Ethereum light crawler

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

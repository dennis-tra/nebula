![Nebula Logo](./docs/nebula-logo.svg)

# Nebula

[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)
[![go test](https://github.com/dennis-tra/nebula/actions/workflows/pull_request_main.yml/badge.svg)](https://github.com/dennis-tra/nebula/actions/workflows/pull_request_main.yml)
[![readme nebula](https://img.shields.io/badge/readme-Nebula-blueviolet)](README.md)
[![GitHub license](https://img.shields.io/github/license/dennis-tra/nebula)](https://github.com/dennis-tra/nebula/blob/main/LICENSE)
[![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fdennis-tra%2Fnebula&count_bg=%2379C83D&title_bg=%23555555&icon=&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://hits.seeyoufarm.com)

A network agnostic peer crawler and monitor. Nebula starts with a set of bootstrap peers, 
asks them for other peers in the network and recursively repeats the process until all
peers in the network have been contacted. Originally, Nebula only supported DHT
networks, but this restriction was lifted.

Currently, Nebula supports the following networks:

- [IPFS](https://ipfs.network) - [_Amino DHT_](https://blog.ipfs.tech/2023-09-amino-refactoring/)
- [Bitcoin](https://bitcoin.org/) | [Litecoin](https://litecoin.org/) | [Dogecoin](https://dogecoin.com/) (alpha)
- [Ethereum](https://ethereum.org/en/) - [_Consensus Layer (discv5)_](https://ethereum.org/uz/developers/docs/networking-layer/#consensus-discovery) | [_Execution Layer (discv4)_](https://ethereum.org/uz/developers/docs/networking-layer/#discovery)
- [Optimism](https://www.optimism.io/) compatible chains
- [Portal](https://www.portal.network/) - (_alpha - [wire protocol](https://github.com/ethereum/portal-network-specs/blob/master/portal-wire-protocol.md) not implemented_)
- [Filecoin](https://filecoin.io)
- [Polkadot](https://polkadot.network/) - [_Kusama_](https://kusama.network/) | [_Rococo_](https://substrate.io/developers/rococo-network/) | [_Westend_](https://wiki.polkadot.network/docs/maintain-networks#westend-test-network)
- [Avail](https://www.availproject.org/) - [_Mainnet_](https://docs.availproject.org/docs/networks#mainnet) | [_Turing_](https://docs.availproject.org/docs/networks#turing-testnet) | _<small>Light Client + Full Node versions</small>_
- [Celestia](https://celestia.org/) - [_Mainnet_](https://blog.celestia.org/celestia-mainnet-is-live/) | [_Mocha_](https://docs.celestia.org/nodes/mocha-testnet) | [_Arabica_](https://github.com/celestiaorg/celestia-node/blob/9c0a5fb0626ada6e6cdb8bcd816d01a3aa5043ad/nodebuilder/p2p/bootstrap.go#L40)
- [Pactus](https://pactus.org)
- [Dria](https://dria.co/)
- [Gnosis](https://www.gnosis.io/)
- ... your network? Get in touch [team@probelab.io](team@probelab.io).

You can run `nebula networks` to get a list of all supported networks

Nebula supports the following storage backends: JSON, Postgres, ClickHouse ([TCP](https://clickhouse.com/docs/interfaces/tcp) protocol)


The crawler was:

- 🏆 _awarded a prize in the [DI2F Workshop hackathon](https://research.protocol.ai/blog/2021/decentralising-the-internet-with-ipfs-and-filecoin-di2f-a-report-from-the-trenches/)._ 🏆
- 🎓 _used for the ACM SigCOMM'22 paper [Design and Evaluation of IPFS: A Storage Layer for the Decentralized Web](https://research.protocol.ai/publications/design-and-evaluation-of-ipfs-a-storage-layer-for-the-decentralized-web/trautwein2022.pdf)_ 🎓

Nebula powers:

- 📊 _the weekly reports for the IPFS Amino DHT [here](https://github.com/probe-lab/network-measurements/tree/main/reports)!_ 📊
- 🌐 _many graphs on [probelab.io](https://probelab.io) for most of the supported networks above_ 🌐

You can find a demo on YouTube: [Nebula: A Network Agnostic DHT Crawler](https://www.youtube.com/watch?v=QDgvCBDqNMc) 📺

![Screenshot from a Grafana dashboard](./docs/grafana-screenshot.png)

<small>_Grafana Dashboard is not part of this repository_</small>

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
- [Demo](#demo)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [Support](#support)
- [Other Projects](#other-projects)
- [License](#license)

## Project Status

The crawler is powering critical [IPFS](https://ipfs.tech) [Amino DHT](https://blog.ipfs.tech/2023-09-amino-refactoring/) [KPIs](https://de.wikipedia.org/wiki/Key-Performance-Indicator), used for [Weekly IPFS Reports](https://github.com/probe-lab/network-measurements/tree/main/reports) as well as for many metrics on [`probelab.io`](https://probelab.io).
The `main` branch will contain the latest changes and should not be considered stable. The latest stable release that is production ready is version [2.4.0](https://github.com/dennis-tra/nebula/releases/tag/2.4.0).

## Install

### Precompile Binaries

Head over to the release section and download binaries from the [latest stable release](https://github.com/dennis-tra/nebula/releases).

### From source

```shell
git clone https://github.com/dennis-tra/nebula
cd nebula
just build
```

Now you should find the `nebula` executable in the `dist` subfolder.

## Usage

Nebula is a command line tool and provides the `crawl` sub-command.

### Dry-Run

To simply crawl the IPFS Amino DHT network run:

```shell
nebula --dry-run crawl
```

> [!NOTE]
> For backwards compatibility reasons IPFS is the default if no network is specified

The crawler can store its results as JSON documents, in a [Postgres](https://www.postgresql.org/), or in a [Clickhouse](https://clickhouse.com/) database -
the `--dry-run` flag prevents it from doing any of it. Nebula will just print a
summary of the crawl at the end instead. For the IPFS network, a crawl takes ~5-10 min depending on
your internet connection. You can also specify the network you want to crawl by
appending, e.g., `--network FILECOIN` and limit the number of peers to crawl by
providing the `--limit` flag with the value of, e.g., `1000`. Example:

```shell
nebula --dry-run crawl --network FILECOIN --limit 1000
```

To find out which other network values are supported, you can run:

```shell
nebula networks
```

### JSON Output

To store crawl results as JSON files provide the `--json-out` command line flag like so:

```shell
nebula --json-out ./results/ crawl
```

After the crawl has finished, you will find the JSON files in the `./results/` subdirectory.

When providing only the `--json-out` command line flag you will see that the
`*_neighbors.json` document is empty. This document would contain the full
routing table information of each peer in the network which is quite a bit of
data (~250MB for the Amino DHT as of April '23) and is therefore disabled by
default

### Track Routing Table Information

To populate the document, you'll need to pass the `--neighbors` flag to
the `crawl` subcommand.

```shell
nebula --json-out ./results/ crawl --neighbors
```

The routing table information forms a graph and graph visualization tools often
operate with [adjacency lists](https://en.wikipedia.org/wiki/Adjacency_list). To convert the `*_neighbors.json` document
to an adjacency list, you can use [`jq`](https://stedolan.github.io/jq/) and the following command:

```shell
jq -r '.NeighborIDs[] as $neighbor | [.PeerID, $neighbor] | @csv' ./results/2025-02-16T14:32_neighbors.json > ./results/2025-02-16T14:32_neighbors.csv
```

### Postgres

If you want to store the information in a proper database, you could run `just start-postgres` to start a local postgres instance via docker in the background and run Nebula like:

```shell
nebula --db-user nebula_local --db-name nebula_local crawl --neighbors
```

At this point, you can also start Nebula's monitoring process, which would periodically probe the discovered peers to track their uptime. Run in another terminal:

```shell
nebula --db-user nebula_local --db-name nebula_local monitor
```

When Nebula is configured to store its results in a postgres database, then it also tracks session information of remote peers. A session is one continuous streak of uptime (see below).

However, this is not implemented for all supported networks. The [ProbeLab](https://probelab.network) team is using the monitoring feature for the IPFS, Celestia, Filecoin, and Avail networks. Most notably, the Ethereum discv4/discv5 and Bitcoin monitoring implementations still need work.

---

There are a few more command line flags that are documented when you run`nebula --help` and `nebula crawl --help`:

## How does it work?

### `crawl`

The `crawl` sub-command starts by connecting to a set of bootstrap nodes and then
requesting the information of other peers in the network using the network-native
discovery protocol. For most supported networks these are several Kademlia
`FIND_NODE` RPCs. For Bitcoin-related networks it's a `getaddr` RPC.

For Kademlia-based networks Nebula constructs the routing tables (kademlia _k_-buckets)
of the remote peer based on its [`PeerID`](https://docs.libp2p.io/concepts/peer-id/). Then `nebula` builds
random `PeerIDs` with common prefix lengths (CPL) that fall in each of the peers' buckets, and asks if it knows any peers that are
closer (XOR distance) to the ones `nebula` just generated. This will effectively yield a list of all `PeerIDs` that a peer has
in its routing table. The process repeats for all found peers until `nebula` does not find any new `PeerIDs`.

> [!TIP]
> You can use the `crawl` sub-command with the global `--dry-run` option that skips any database operations.

Command line help page:

```text
NAME:
   nebula crawl - Crawls the entire network starting with a set of bootstrap nodes.

USAGE:
   nebula crawl [command options]

OPTIONS:
   --addr-dial-type value                               Which type of addresses should Nebula try to dial (private, public, any) (default: "public") [$NEBULA_CRAWL_ADDR_DIAL_TYPE]
   --addr-track-type value                              Which type addresses should be stored to the database (private, public, any) (default: "public") [$NEBULA_CRAWL_ADDR_TRACK_TYPE]
   --bootstrap-peers value [ --bootstrap-peers value ]  Comma separated list of multi addresses of bootstrap peers (default: default IPFS) [$NEBULA_CRAWL_BOOTSTRAP_PEERS, $NEBULA_BOOTSTRAP_PEERS]
   --limit value                                        Only crawl the specified amount of peers (0 for unlimited) (default: 0) [$NEBULA_CRAWL_PEER_LIMIT]
   --neighbors                                          Whether to persist all k-bucket entries of a particular peer at the end of a crawl. (default: false) [$NEBULA_CRAWL_NEIGHBORS]
   --network nebula networks                            Which network should be crawled. Presets default bootstrap peers and protocol. Run: nebula networks for more information. (default: "IPFS") [$NEBULA_CRAWL_NETWORK]
   --protocols value [ --protocols value ]              Comma separated list of protocols that this crawler should look for [$NEBULA_CRAWL_PROTOCOLS, $NEBULA_PROTOCOLS]
   --workers value                                      How many concurrent workers should dial and crawl peers. (default: 1000) [$NEBULA_CRAWL_WORKER_COUNT]

   Network Specific Configuration:

   --check-exposed               IPFS/AMINO: Whether to check if the Kubo API is exposed. Checking also includes crawling the API. (default: false) [$NEBULA_CRAWL_CHECK_EXPOSED]
   --keep-enr                    ETHEREUM_CONSENSUS: Whether to keep the full ENR. (default: false) [$NEBULA_CRAWL_KEEP_ENR]
   --udp-response-timeout value  ETHEREUM_EXECUTION: The response timeout for UDP requests in the disv4 DHT (default: 3s) [$NEBULA_CRAWL_UDP_RESPONSE_TIMEOUT]

```

### `monitor`

The `monitor` sub-command is only implemented for libp2p based networks and with the postgres database backend.
It polls every 10 seconds all sessions from the database (see above) that are due to be dialed
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
   nebula monitor [command options]

OPTIONS:
   --workers value  How many concurrent workers should dial peers. (default: 1000) [$NEBULA_MONITOR_WORKER_COUNT]
   --network value  Which network belong the database sessions to. Relevant for parsing peer IDs and muti addresses. (default: "IPFS") [$NEBULA_MONITOR_NETWORK]
   --help, -h       show help

```

### `resolve`

The resolve sub-command is only available when using the postgres datbaase backend. It goes through all multi addresses that are present in the database and resolves them to their respective IP-addresses. Behind one multi address can be multiple IP addresses due to, e.g., the [`dnsaddr` protocol](https://github.com/multiformats/multiaddr/blob/master/protocols/DNSADDR.md).
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

To develop this project, you need Go `1.23` and the following tools:
```
github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2
github.com/volatiletech/sqlboiler/v4@v4.18.0
github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.18.0
go.uber.org/mock/mockgen@v0.5.0
mvdan.cc/gofumpt@v0.7.0
```

To install the necessary tools you can run `just tools`. This will use the `go install` command to download and install the tools into your `$GOPATH/bin` directory. So make sure you have it in your `$PATH` environment variable.

### Database

You need a running Postgres or ClickHouse instance to persist and/or read the crawl results.
Run `just start-postgres` or `just start-clickhouse` respectively
or use one of the following commands:

```shell
# for postgres
docker run --rm -d --name nebula-postgres-local -p 5432:5432 -e POSTGRES_DB=nebula_local -e POSTGRES_USER=nebula_local -e POSTGRES_PASSWORD=password_local postgres:14

# for clickhouse
docker run --rm -d --name nebula-clickhouse-local -p 8123:8123 -p 9000:9000 -e CLICKHOUSE_DB=nebula_local -e CLICKHOUSE_USER=nebula_local -e CLICKHOUSE_PASSWORD=password_local clickhouse/clickhouse-server:24.12
```

Then you can connect to the database with:

```shell
just repl-postgres
# or
just repl-clickhouse
```

To stop the containers:

```shell
just stop-postgres
# or
just stop-clickhouse
```

for convenience there are also the `just restart-postgres` and `just restart-clickhouse` recipes.

> [!TIP]
> You can use the `crawl` sub-command with the global `--dry-run` option that skips any database operations or store the results as JSON files with the `--json-out` flag.

The default database settings for local development are:

```toml
Name     = "nebula_local"
Password = "password_local"
User     = "nebula_local"
Host     = "localhost"
# postgres
Port     = 5432
# clickhouse
Port     = 9000
```

Migrations are applied automatically when `nebula` starts and successfully establishes a database connection.

To run them manually you can run:

```shell
# Up migrations
just migrate-postgres up
just migrate-clickhouse up

# Down migrations
just migrate-postgres down
just migrate-clickhouse down

# Generate the ORMs with SQLBoiler (only postgres)
just models # runs: sqlboiler
```

```shell
# Create new migration
# postgres
migrate create -ext sql -dir db/migrations/pg -seq some_migration_name

# clickhouse
migrate create -ext sql -dir db/migrations/chlocal -seq some_migration_name
```

> [!NOTE]
> Make sure to adjust the `chlocal` migration and copy it over to the `chcluster` folder. In a clustered clickhouse deployment the table engines need to be prefixed with `Replicated`, like `ReplicatedMergeTree` as opposed to just `MergeTree`.

### Tests

To run the tests you need a running test database instance. The following command
starts test postgres and clickhouse containers, runs the tests and tears them
down again:

```shell
just test
```

The test database containers won't interfere with other local containers as
all names etc. are suffixed with `_test` as opposed to `_local`.

To speed up running database tests you can do the following:

```shell
just start-postgres test
just start-clickhouse test
```

Then run the plain tests (without starting database containers):
```shell
just test-plain
```

Eventually, stop the containers again:

```shell
just stop-postgres test
just stop-clickhouse test
```

## Release Checklist

- [ ] Merge everything into `main`
- [ ] Create a new tag with the new version
- [ ] Push tag to GitHub

This will trigger the [`goreleaser.yml`](./.github/workflows/goreleaser.yml) workflow which pushes creates a new _draft_ release in GitHub.

## Related Efforts

- [`wiberlin/ipfs-crawler`](https://github.com/wiberlin/ipfs-crawler) - A crawler for the IPFS network, code for their paper ([arXiv](https://arxiv.org/abs/2002.07747)).
- [`adlrocha/go-libp2p-crawler`](https://github.com/adlrocha/go-libp2p-crawler) - Simple tool to crawl libp2p networks resources
- [`libp2p/go-libp2p-kad-dht`](https://github.com/libp2p/go-libp2p-kad-dht/tree/master/crawler) - Basic crawler for the Kademlia DHT implementation on go-libp2p.
- [`migalabs/armiarma`](https://github.com/migalabs/armiarma) - Armiarma is a Libp2p open-network crawler with a current focus on Ethereum's CL network
- [`migalabs/eth-light-crawler`](https://github.com/migalabs/eth-light-crawler) - Ethereum light crawler by [@cortze](https://github.com/cortze).

## Demo

The following presentation shows a ways to use Nebula by showcasing crawls of the Amino, Celestia, and Ethereum DHT's:

[![Nebula: A Network Agnostic DHT Crawler - Dennis Trautwein](https://img.youtube.com/vi/QDgvCBDqNMc/0.jpg)](https://www.youtube.com/watch?v=QDgvCBDqNMc)

## Networks

> [!NOTE]
> This section is work-in-progress and doesn't include information about all networks yet.

The following sections document our experience with crawling the different networks.

### Ethereum Execution (disv4)

Under the hood Nebula uses packages from [`go-ethereum`](https://github.com/ethereum/go-ethereum) to facilitate peer
communication. Mostly, Nebula relies on the [discover package](https://github.com/ethereum/go-ethereum/tree/master/p2p/discover).
However, we made quite a few changes to the implementation that can be found in
our fork of `go-ethereum` [here](https://github.com/probe-lab/go-ethereum/tree/nebula) in the `nebula` branch.

Most notably, the custom changes include:

- export of internal constants, functions, methods and types to customize their behaviour or call them directly
- changes to the response matcher logic. UDP packets won't be forwarded to all matchers. This was required so that
  concurrent requests to the same peer don't lead to unhandled packets

Deployment recommendations:

- CPUs: 4 (better 8)
- Memory > 4 GB
- UDP Read Buffer size >1 MiB (better 4 MiB) via the `--udp-buffer-size=4194304` command line flag or corresponding environment variable `NEBULA_UDP_BUFFER_SIZE`.
  You might need to adjust the maximum buffer size on Linux, so that the flag takes effect:
  ```shell
  sysctl -w net.core.rmem_max=8388608 # 8MiB
  ```
- UDP Response timeout of `3s` (default)
- Workers: 3000

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

## License

[Apache License Version 2.0](LICENSE) © Dennis Trautwein

![Nebula Crawler Logo](./docs/nebula-logo.svg)

# Nebula Crawler

[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)
[![readme nebula](https://img.shields.io/badge/readme-Nebula-blueviolet)](README.md)
[![GitHub license](https://img.shields.io/github/license/dennis-tra/nebula-crawler)](https://github.com/dennis-tra/nebula-crawler/blob/main/LICENSE)

A libp2p DHT crawler that gathers information about running nodes in the network. The crawler runs every 30 minutes by connecting to the standard DHT bootstrap nodes and then recursively following all entries in the k-buckets until all peers have been visited.

<details><summary>Screenshot</summary>
<img src="./docs/grafana-screenshot.png" alt="Screenshot from a Grafana dashboard">
</details>

## Table of Contents

- [Project Status](#project-status)
- [Usage](#usage)
- [How does it work?](#how-does-it-work)
  - [`crawl`](#crawl) | [`monitor`](#monitor) | [`daemon`](#daemon)
- [Install](#install)
  - [Release download](#release-download) | [From source](#from-source)
- [Development](#development)
  - [Database](#database)
- [Related Efforts](#related-efforts)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [Support](#support)
- [Other Projects](#other-projects)
- [License](#license)

[comment]: <> (## Motivation)

[comment]: <> (https://github.com/protocol/ResNetLab/discussions/34)

## Project Status

The crawler is successfully visiting and following all reachable nodes in the IPFS and Filecoin networks. However, the project is still very young and thus has its sharp edges here and there. Most importantly, the gathered numbers about the IPFS network are in line with existing data like from the [`wiberlin/ipfs-crawler`](https://github.com/wiberlin/ipfs-crawler). Their crawler also powers a dashboard which can be found [here](https://trudi.weizenbaum-institut.de/ipfs_analysis.html).

## Usage

Nebula is a command line tool and provides the three sub-commands `crawl`, `monitor` and `daemon`. To simply crawl the IPFS network run:

```shell
nebula crawl --dry-run
```

The `--dry-run` flag does not persist any data. One run takes ~5-10 min dependent on your internet connection.

See the command line help page below for configuration options:

```shell
NAME:
   nebula - A libp2p DHT crawler and monitor that exposes timely information about DHT networks.

USAGE:
   nebula [global options] command [command options] [arguments...]

VERSION:
   vdev+5f3759df

AUTHOR:
   Dennis Trautwein <nebula@dtrautwein.eu>

COMMANDS:
   crawl    Crawls the entire network based on a set of bootstrap nodes.
   monitor  Monitors the network by periodically dialing and pinging previously crawled peers.
   daemon   Start a long running process that crawls and monitors the DHT network
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug                  Set this flag to enable debug logging (default: false) [$NEBULA_DEBUG]
   --log-level value        Set this flag to a value from 0 to 6. Overrides the --debug flag (default: 4) [$NEBULA_LOG_LEVEL]
   --config FILE            Load configuration from FILE [$NEBULA_CONFIG_FILE]
   --dial-timeout value     How long should be waited before a dial is considered unsuccessful (default: 30s) [$NEBULA_DIAL_TIMEOUT]
   --prom-port value        On which port should prometheus serve the metrics endpoint (default: 6666) [$NEBULA_PROMETHEUS_PORT]
   --prom-host value        Where should prometheus serve the metrics endpoint (default: localhost) [$NEBULA_PROMETHEUS_HOST]
   --db-host value          On which host address can nebula reach the database (default: localhost) [$NEBULA_DATABASE_HOST]
   --db-port value          On which port can nebula reach the database (default: 5432) [$NEBULA_DATABASE_PORT]
   --db-name value          The name of the database to use (default: nebula) [$NEBULA_DATABASE_NAME]
   --db-password value      The password for the database to use (default: password) [$NEBULA_DATABASE_PASSWORD]
   --db-user value          The user with which to access the database to use (default: nebula) [$NEBULA_DATABASE_USER]
   --protocols value        Comma separated list of protocols that this crawler should look for (default: "/ipfs/kad/1.0.0", "/ipfs/kad/2.0.0") [$NEBULA_PROTOCOLS]
   --bootstrap-peers value  Comma separated list of multi addresses of bootstrap peers [$NEBULA_BOOTSTRAP_PEERS]
   --help, -h               show help (default: false)
   --version, -v            print the version (default: false)
```

## How does it work?

### `crawl`

The `crawl` sub-command starts by connecting to a set of bootstrap nodes and constructing the routing tables (kademlia _k_-buckets)
of the remote peers based on their [`PeerIds`](https://docs.libp2p.io/concepts/peer-id/). Then `nebula` builds
random `PeerIds` with a common prefix length (CPL) and asks each remote peer if they know any peers that are
closer to the ones `nebula` just constructed. This will effectively yield a list of all `PeerIds` that a peer has
in its routing table. The process repeats for all found peers until `nebula` does not find any new `PeerIds`.

This process is heavily inspired by the `basic-crawler` in [libp2p/go-libp2p-kad-dht](https://github.com/libp2p/go-libp2p-kad-dht/tree/master/crawler) from @aschmahmann.

Every peer that was found is persisted together with its multi-addresses. If the peer was dialable `nebula` will
also create a `session` instance that contains the following information:

```go
type Session struct {
  // A unique id that identifies a particular session
  ID int
  // The peer ID in the form of Qm... or 12D3...
  PeerID string
  // When was the peer successfully dialed the first time
  FirstSuccessfulDial time.Time
  // When was the most recent successful dial to the peer above
  LastSuccessfulDial time.Time
  // When should we try to dial the peer again
  NextDialAttempt null.Time
  // When did we notice that this peer is not reachable.
  // This cannot be null because otherwise the unique constraint
  // uq_peer_id_first_failed_dial would not work (nulls are distinct).
  // An unset value corresponds to the timestamp 1970-01-01
  FirstFailedDial time.Time
  // The duration that this peer was online due to multiple subsequent successful dials
  MinDuration null.String
  // The duration from the first successful dial to the point were it was unreachable
  MaxDuration null.String
  // indicates whether this session is finished or not. Equivalent to check for
  // 1970-01-01 in the first_failed_dial field.
  Finished bool
  // How many subsequent successful dials could we track
  SuccessfulDials int
  // When was this session instance updated the last time
  UpdatedAt time.Time
  // When was this session instance created
  CreatedAt time.Time
}
```

At the end of each crawl `nebula` persists general statistics about the crawl like the duration, dialable peers, encountered errors, agent versions etc...

> **Info:** You can use the `crawl` sub-command with the `--dry-run` option that skips any database operations.

### `monitor`

The `monitor` sub-command polls every 10 seconds all sessions from the database (see above) that are due to be dialed
in the next 10 seconds (based on the `NextDialAttempt` timestamp). It attempts to dial all peers using previously
saved multi-addresses and updates their `session` instances accordingly if they're dialable or not.

The `NextDialAttempt` timestamp is calculated based on the uptime that `nebula` has observed for that given peer.
If the peer is up for a long time `nebula` assumes that it stays up and thus decreases the dial frequency aka. sets
the `NextDialAttempt` timestamp to a time further in the future.

### `daemon`

**Work in progress:** The `daemon` sub-command combines the `crawl` and `monitor` tasks in a single process. It uses application level
scheduling of the crawls rather than e.g. using OS-level cron configurations.

## Install

### Release download

There is no release yet.

### From source

To compile it yourself run:

```shell
go install github.com/dennis-tra/nebula/cmd/nebula@latest # Go 1.16 or higher is required (may work with a lower version too)
```

Make sure the `$GOPATH/bin` is in your PATH variable to access the installed `nebula` executable.

## Development

To develop this project you need Go `> 1.16` and the following tools:

- [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) to manage the SQL migration `v4.14.1`
- [`volatiletech/sqlboiler`](https://github.com/volatiletech/sqlboiler) to generate Go ORM `v4.6.0`

To install the necessary tools you can run `make tools`. This will use the `go install` command to download and install the tools into your `$GOPATH/bin` directory. So make sure you have it in your `$PATH` environment variable.

### Database

You need a running postgres instance to persist and/or read the crawl results. Use the following command to start a local instance of postgres:

```shell
docker run -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=nebula -e POSTGRES_DB=nebula postgres:13
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
migrate -database 'postgres://nebula:password@localhost:5432/nebula?sslmode=disable' -path migrations up
# OR
make migrate-up

# Down migrations
migrate -database 'postgres://nebula:password@localhost:5432/nebula?sslmode=disable' -path migrations down
# OR
make migrate-down

# Create new migration
migrate create -ext sql -dir migrations -seq some_migration_name
```

To generate the ORM with SQLBoiler run:

```shell
sqlboiler psql
```

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

## License

[Apache License Version 2.0](LICENSE) © Dennis Trautwein

# ⏱ IPFS Uptime Measurement Summary
### _Objective 3 as described in [here](https://docs.google.com/document/d/1ap6TVdI2AHAllt0cj_Zv7r_25bTA7k2zY873Dg-7r7Q/edit#)_

## >>>>Summary for storm nodes analysis>>>>
This section is a summary for storm nodes. If you want to know the crawling summary, please start from the next section.
Storm node is a type of node that is built on top of libp2p. Storm nodes aim to attack unix systems. We built a monitor that can be used to interact with storm nodes cralwed. The way of interaction is that after we connected to storm nodes, we will try to keep letting them store a random provider record and fetch later. In order to run the monitor. Follow the instructions below:
```
cd analysis/mixed
python3 get_storm_nodes.py
cd ../../cmd/storm-monitor
go run main.go
```
After the monitoring, we found out that most storm nodes participate in the DHT Network of IPFS. That is, they are willing to store and fetch provider record. Besides that, they seem to not have interactions with other nodes in IPFS. However, from this article [here](https://www.bitdefender.com/files/News/CaseStudies/study/376/Bitdefender-Whitepaper-IPStorm.pdf), storm nodes communicate using custom libp2p protocols between each other such as `sbpcp/1.0.0` and `sbptp/1.0.0`. These custom protocols have also been discovered by the crawler.

We also plotted the response time for the storm nodes. We recorded the response time for putting a provider record and the response time for fetching a provider record. As shown below.

This is the response time for putting a provider record to the storm nodes.

![Graph](./analysis/mixed/figs/storm_put_response_time.png)

We exlude outliers and below is what we got.

![Graph](./analysis/mixed/figs/storm_put_response_time_trim.png)

We also plotted the response time for fetching a provider record from the storm nodes.

![Graph](./analysis/mixed/figs/storm_fetch_response_time.png)

We exlude outliers and below is what we got.

![Graph](./analysis/mixed/figs/storm_fetch_response_time_trim.png)

If you want to plot the same figures. After you run the `go run main.go`, follow the instructions below:
```
cd ../../analysis/mixed
python3 plot_storm_response_time.py
```

## Summary
- We have made some extensions to the original [nebula crawler](https://github.com/dennis-tra/nebula-crawler) so it is able to collect some additional data to analyse the churn. The fork of the repository lives [here](https://github.com/wcgcyx/nebula-crawler).
- We created a list of python functions that interact directly with the database to collect useful data.
- We created a list of python scripts (built on top of the functions) to automatically generate graphs we are interested.
- We collected some sample data and generated some sample graphs.

## Quick Deployment
To quickly deploy the crawler & monitor, follow the procedures below (If you don't have docker installed, checkout [here](https://www.docker.com/)).
```
git clone https://github.com/wcgcyx/nebula-crawler.git
cd nebula-crawler
make docker
cd deploy
docker-compose up
```
This will start a docker compose that contains one container for collecting data and one container for postgres DB. Our scripts are built to request directly to the DB.
You can configure how the crawl works and the DB. The [docker compose file](https://github.com/wcgcyx/nebula-crawler/blob/main/deploy/docker-compose.yml) contains a list of configurations among which that are related to the data collection is as below.
| Env name | Description |
| ------------- |:-------------|
| NEBULA_DATABASE_HOST | Database host |
| NEBULA_DATABASE_PORT | Database port |
| NEBULA_DATABASE_NAME | Database name |
| NEBULA_DATABASE_USER | Database user |
| NEBULA_DATABASE_PASSWORD | Database password |
| NEBULA_CRAWL_PERIOD | The gap between two crawls in seconds. If set to 1800 then the crawler will run every 30 minutes |
| NEBULA_SAVE_NEIGHBOURS | Flag indicates whether this cralwer should persist neighbours data. If set to 1, it will save neighbours. |
| NEBULA_NOT_TRUNCATE_NEIGHBOURS | Flag indicates whether this cralwer should truncate previous neighbours. If set to 0, every time it runs it will truncate pervious neighbours. |
| NEBULA_SAVE_CONNECTIONS | Flag indicates whether this cralwer should persist libp2p connect latency data. If set to 1, it will save latency. |
| NEBULA_NOT_TRUNCATE_CONNECTIONS | Flag indicates whether this cralwer should truncate previous connections. If set to 0, every time it runs it will truncate pervious connections. |
Note that every crawl will produce roughly 350MB neighbours data and 15MB connections data so it is recommended to truncate both data to limit DB growth.

## System requirement
The plotting script in theory should not consume more than 500MB of memory. However, if you decide to run the docker-compose and script locally at the same time, it is recommended you run the script with the following specs:

CPU: 6-Core 2.6GHz or better

Memory: 16GB Ram or more

## Quick Data Analysis
To quickly generate some graphs, you will need Python 3, if you don't have Python 3 installed, checkout [here](https://www.python.org/).
```
cd nebula-crawler/analysis/mixed
python3 ./plot_churn.py             # This will plot the network churn.
python3 ./plot_nodes.py             # This will plot a pie chart of on node, off node and dangling node.
python3 ./plot_agent_all.py         # This will plot the agent versions across all the nodes.
python3 ./plot_protocol_all.py      # This will plot the protocols across all the nodes.
python3 ./plot_geo_all.py           # This will plot the geolocations across all the nodes.
python3 ./plot_cloud_all.py         # This will plot the cloud information across all the nodes.
```
We have generated some graphs using the scripts above to `analysis/mixed/figs` and there are some more scripts that generate different plots. One thing to notice is that you will need root access to run icmp ping related scripts. What is most important is that you can easily extend these plotting scripts by piping between the libraires so that it will plot information that you are interested in. We have included a few libraries, under the `analysis/mixed/lib/` and are described below.
| Library file | Description |
| ------------- |:-------------|
| `node_agent.py` | This library returns the agent version of given peer IDs. |
| `node_classification.py` | This library returns the node classification of given peer IDs. |
| `node_cloud.py` | This library returns the cloud information of given peer IDs. |
| `node_correlation.py` | This library returns the correlation between the up time and the local day time of given peer IDs. |
| `node_geolocation.py` | This library returns the geolocation of given peer IDs.  |
| `node_latency.py` | This library returns the libp2p connection latency of given peer IDs. |
| `node_ping.py` | This library returns if ICMP Ping works to given peer IDs. |
| `node_protocol.py` | This library returns the protocol of given peer IDs. |
| `node_reliability.py` | This library returns the reliability of given peer IDs. |
| `node_time.py` | This library returns the min DB timestamp and the max DB timestamp for use when calling other libraries. |
| `node_uptime.py` | This library returns the uptime of given peer IDs. |
If you are unsure about how to use these libraries, please have a look at the plotting scripts, which are examples of using the libraries. In addition, it is entirely possible to come up with new libraries, please have a look at the database schemas [here](https://github.com/wcgcyx/nebula-crawler/tree/main/migrations).

## Terminology
| Term | Description |
| ------------- |:-------------|
| On node | On node refers to a type of IPFS nodes that are always on since the first time we discovered it. |
| Off node | Off node refers to a type of IPFS nodes that we never got a successful connection to since the first time we discovered it. |
| Dangling node | Dangling node refers to a type of IPFS nodes that are only sometimes reachable since the first time we discovered it. |
| Node reliability | The percentage of uptime over the total time. |

## Results
In this section, we will present some graphs we collected using the crawler and the scripts.
### Whole network
First we generated some graph on the whole network.

#### 1 Network churn  
![Graph](./analysis/mixed/figs/network_churn.png)  
Around 50% of peers sray in the network for 1.5 hour or less and around 70% of peers stay in the network for 3 hours or less.

#### 2 Node classification  
![Graph](./analysis/mixed/figs/node_classification.png)  
Majority of the nodes are dangling nodes (>85%), with 14% being the on nodes.

#### 3 Agent version of **all the nodes**  
![Graph](./analysis/mixed/figs/agent_version_for_all_nodes.png)  
We did an analysis of the agent version of all the nodes and discovered that majority of the nodes are running go-ipfs, with go-ipfs 0.8.x being the most popular. If we break down the 0.8.x further, the most popular tag is `48f94e2`.

#### 4 Protocol of **all the nodes**  
![Graph](./analysis/mixed/figs/protocols_for_all_nodes.png)

#### 5 Geolocation of **all the nodes**  
![Graph](./analysis/mixed/figs/geolocation_for_all_nodes.png)

#### 6 Cloud information of **all the nodes**  
![Graph](./analysis/mixed/figs/cloud_info_for_all_nodes.png)  
As shown in the graph, 15% of all the IPFS nodes are running in cloud. Digital Ocean is the most popular cloud provider, used by 10.5% of the nodes.
### On nodes
We then pick the on nodes (nodes that are always on) and analyse them further.

#### 1 Agent version of **on nodes**  
![Graph](./analysis/mixed/figs/agent_version_for_on_nodes.png)  
Over 90% of the on nodes are running go-ipfs, higher than network (77%).

#### 2 Protocol of **on nodes**  
![Graph](./analysis/mixed/figs/protocols_for_on_nodes.png)

#### 3 Geolocation of **on nodes**  
![Graph](./analysis/mixed/figs/geolocation_for_on_nodes.png)  
It is interesting to see that when comes to reliable nodes, 18% of them are running in Germany.

#### 4 Cloud information of **on nodes**  
![Graph](./analysis/mixed/figs/cloud_info_for_on_nodes.png)  
When it comes to cloud information, unsuprisingly that over 25% of the on nodes are running in a cloud environment, higher than network average (15%). Digital Ocean and AWS are widely used (each taking around 11%).
### Dangling nodes
We finally analyse the dangling nodes (nodes that are sometimes on/restarted during monitoring).

#### 1 Agent version of **dangling nodes**  
![Graph](./analysis/mixed/figs/agent_version_for_dangling_nodes.png)  
The agent of dangling nodes are similar to the whole network.

#### 2 Protocol of **dangling nodes**  
![Graph](./analysis/mixed/figs/protocols_for_dangling_nodes.png)

#### 3 Geolocation of **dangling nodes**  
![Graph](./analysis/mixed/figs/geolocation_for_dangling_nodes.png)  
The geolocation of dangling nodes are also similar to the whole network.

#### 4 Cloud information of **dangling nodes**  
![Graph](./analysis/mixed/figs/cloud_info_for_dangling_nodes.png)  
12% of the dangling nodes are running in a cloud environment, a little bit lower than network average.

#### 5 Reliability  
![Graph](./analysis/mixed/figs/dangling_nodes_reliability.png)  
This reliability graph also indirectly shows the network churn, majority of the nodes are only accessible <5% of the time.

#### 6 Correlation between uptime and geolocation  
![Graph](./analysis/mixed/figs/dangling_nodes_hk_uptime_daytime_correlation.png)  
We analysed dangling nodes that are in Hong Kong as an example. We calculated the correlation between the uptime of a node and the local day time and we also calculated the correlation between the uptime of a node and the local night time to compare. Unfortunately, there is no clear signal showing that the uptime is correlated with the local day time. If there is strong correlation between the uptime and the local day time, we would have seen more counts of high correlation and the two graphs will have different distribution. However, what we see here is that most correlations are close to 0 and the two graphs are of similar distribution (meaning that the day time or night time does not matter too much).
Below is an example on how we calculate the correlation:
```
uptime = [0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0] # 1 means it is up at that timestamp, 0 means down and every element in array represents a few minutes slot.
daytime = [0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0] # 1 means it is daytime at that timestamp, 0 means night time and every element in array represents the same slot as above.
corr = numpy.correlate(uptime, daytime) # Result is 2.
```
### Plots from large dataset
Thanks to the original author of nebula crawler Dennis ([@dennis-tra](https://github.com/dennis-tra/)) to provide a large dataset from a 2 months data collection. We are not able to generate all of the plots above because we build some of the plots on extended schema. Here is what we generated from the large dataset.  
#### 1 Network churn  
![Graph](./analysis/mixed/figs/network_churn_d.png)

#### 2 Cloud information of *all seen nodes*  
![Graph](./analysis/mixed/figs/cloud_info_for_all_nodes_d.png)

#### 3 Geolocation of *all seen nodes*  
![Graph](./analysis/mixed/figs/geolocation_for_all_nodes_d.png)  
One interesting thing is that among all seen nodes, there are 52.3% nodes running from Digital Ocean. However, this does not mean that 50% of the nodes are running in the cloud, it is likely that IPFS running in Digital Ocean can restart with a different peer ID and thus contributed to one count in the pie chart.

### Plots for storm nodes
#### 1 Storm nodes classification
![Graph](./analysis/mixed/figs/storm_node_classification.png)

#### 2 Storm nodes geolocation
![Graph](./analysis/mixed/figs/geolocation_for_all_storm.png)

#### 3 Storm nodes cloud infomation
![Graph](./analysis/mixed/figs/cloud_info_for_all_storm.png)

### Dangling nodes off time

#### 1 CDF of total off time for dangling nodes
![Graph](./analysis/mixed/figs/total_offtime_cdf_dangle.png)

#### 2 CDF of off time for dangling nodes
![Graph](./analysis/mixed/figs/offtime_cdf_dangle.png)

### Latency map
We deployed a few nodes in different regions in AWS. We plotted a few graphs in different regions and they can be seen under `analysis/mixed/region-figs`. We also plotted a latency map as shown below.
![Graph](./analysis/mixed/region-latencies/latency-map.png)
Every line has 9 possible colors: dark green, green, yellow, orange, sandy brown, orange red, red, dark red, maroon. Dark green is used if the latency between these two points is less than 10% of the max latency of all data. Similarly, for green, it is 10% - 20%. Yellow (20% - 30%). Orange (30% - 40%). Sandy brown (40% - 50%). Orange read (50% - 60%). Red (60% - 70%). Dark red (70% - 80%). Maroon (80%-100%). You can easily modify the line color by chaning lines at [here](https://github.com/wcgcyx/nebula-crawler/blob/812f33515342321461e39b17ac02a91858926e14/analysis/mixed/plot_map.py#L16-L50).

If you would like to plot the same latency map. First of all, you need to download the latency data. For example,
```
cd analysis/mixed/
python3 ./plot_latency > ./region-latencies/london.latency
python3 plot_map.py
```
To deploy a few nodes in AWS. First of all, you need to make sure you have terraform cli and AWS cli installed, check https://learn.hashicorp.com/tutorials/terraform/install-cli and https://aws.amazon.com/cli/ for information on how to install on your machine.
To check if you have them installed:
```
terraform version
aws --version
```
Secondly, you need to configure aws with your credentials by run `aws configure`.
Lastly, you are ready to spin up the nodes. For example, in order to spin up a node in london:
```
cd regions/london
terraform init (Run only at first time)
terraform apply
```

### Conclusion
- We have made some extensions to the original so it is able to collect some additional data to analyse the churn.
- We created a list of python functions that interact directly with the database to collect useful data.
- We created a list of python scripts (built on top of the functions) to automatically generate graphs we are interested.
- We collected some sample data and generated some sample graphs.
- If there is anything you are interested in, feel free to try it out or reach to us. We will try to keep it updated as long as possible. Meanwhile we will shift our focus to objective 4.

### Below are original README
---

![Nebula Crawler Logo](./docs/nebula-logo.svg)

# Nebula Crawler

[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)
[![readme nebula](https://img.shields.io/badge/readme-Nebula-blueviolet)](README.md)
[![GitHub license](https://img.shields.io/github/license/dennis-tra/nebula-crawler)](https://github.com/dennis-tra/nebula-crawler/blob/main/LICENSE)

A libp2p DHT crawler that also monitors the liveness and availability of peers. The crawler runs every 30 minutes by connecting to the standard DHT bootstrap nodes and then recursively following all entries in the k-buckets until all peers have been visited. Currently I'm running it for the IPFS and Filecoin networks.

🏆 _The crawler was awarded a prize in the context of the [DI2F Workshop hackathon](https://research.protocol.ai/blog/2021/decentralising-the-internet-with-ipfs-and-filecoin-di2f-a-report-from-the-trenches/)._ 🏆

![Screenshot from a Grafana dashboard](./docs/grafana-screenshot.png)

## Table of Contents

- [Project Status](#project-status)
- [Usage](#usage)
- [How does it work?](#how-does-it-work)
  - [`crawl`](#crawl) | [`monitor`](#monitor) | [`daemon`](#daemon)
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

The crawler is successfully visiting and following all reachable nodes in the IPFS and Filecoin networks. However, the project is still very young and thus has its sharp edges here and there. Most importantly, the gathered numbers about the IPFS network are in line with existing data like from the [`wiberlin/ipfs-crawler`](https://github.com/wiberlin/ipfs-crawler). Their crawler also powers a dashboard which can be found [here](https://trudi.weizenbaum-institut.de/ipfs_analysis.html).

## Usage

Nebula is a command line tool and provides the three sub-commands `crawl`, `monitor` and `daemon`. To simply crawl the IPFS network run:

```shell
nebula crawl --dry-run
```

Usually the crawler will persist its result in a postgres database - the `--dry-run` flag prevents it from doing that. One run takes ~5-10 min dependent on your internet connection.

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
scheduling of the crawls rather than e.g. OS-level cron configurations.

## Install

### Release download

There is no release yet. 

### From source

To compile it yourself run:

```shell
go install github.com/dennis-tra/nebula-crawler/cmd/nebula@latest # Go 1.16 or higher is required (may work with a lower version too)
```

Make sure the `$GOPATH/bin` is in your PATH variable to access the installed `nebula` executable.

## Development

To develop this project you need Go `> 1.16` and the following tools:

- [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) to manage the SQL migration `v4.14.1`
- [`volatiletech/sqlboiler`](https://github.com/volatiletech/sqlboiler) to generate Go ORM `v4.6.0`
- `docker` to run a local postgres instance

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

## Deployment

First, you need to build the nebula docker image:

```shell
make docker
# OR
docker build . -t dennis-tra/nebula-crawler:latest
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
  --name nebula_crawler \
  --hostname nebula_crawler \
  dennis-tra/nebula-crawler:latest \
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
nebula --prom-port=6667 monitor --workers=1000 # for IPFS
nebula --prom-port=6669 --config filecoin.json monitor --workers=1000 # for Filecoin
```

The `filecoin.json` file contains the following content:

```shell
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
nebula crawl --workers=1000 # for IPFS (uses defaults like prom-port 6666)
nebula --config filecoin.json crawl --workers=1000 # for Filecoin (uses configuration prom-port 6668)
```

The `workers` flag configures the amount of concurrent connection/dials. I increased it until I didn't notice any performance improvement anymore.

So this is the prometheus ports configuration:

- `nebula crawl` - `ipfs` - `6666`
- `nebula monitor` - `ipfs` - `6667`
- `nebula crawl` - `filecoin` - `6668`
- `nebula monitor` - `filecoin` - `6669`

Furthermore, `nebula` has a hidden flag called `--pprof-port` if this flag is set it also serves `pprof` at `localhost:given-port` for debugging.
## Analysis

There is a top-level `analysis` folder that contains various scripts to help understand the gathered data. More information can be found in the respective subfolders README file. The following evaluations can be found there

- [`geoip`](./analysis/geoip) - Uses a [Maxmind](https://maxmind.com) database to map IP addresses to country ISO codes and plots the results.
- [`churn`](./analysis/churn) - Uses a `sessions` database dump to construct a CDF of peer session lengths.
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

## License

[Apache License Version 2.0](LICENSE) © Dennis Trautwein

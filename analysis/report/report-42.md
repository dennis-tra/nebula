# Nebula Measurement Results Calendar Week 42

## General Information

The measurements were conducted on the following machine:

- `vCPU` - `4`
- `RAM` - `8GB`
- `Disk` - `160GB`
- `Datacenter` - `nbg1-dc3`
- `Country` - `Germany`
- `City` - `Nuremberg`

The following results show measurement data that was collected in calendar week 42 from 2021-10-18 to 2021-10-25 in 2021.

- Number of crawls `336`
- Number of visits `8,532,595` ([what is a visit?](#terminology))
- Number of unique peer IDs visited `59,083`
- Number of unique IP addresses found `94,080`

### Agent Versions

Newly discovered agent versions:

- `go-ipfs/0.10.0/b5b5f09b6-dirty` (2021-10-17 23:32:14)
- `go-ipfs/0.9.1/d841f42bb` (2021-10-18 10:01:34)
- `go-open-p2p` (2021-10-18 13:04:45)
- `RendezvousRAT/server` (2021-10-19 11:31:46)
- `go-ipfs/0.9.1/1b6fb661c` (2021-10-20 07:30:56)
- `go-ipfs/0.11.0-dev/5a61bed` (2021-10-20 17:01:03)
- `go-ipfs/0.11.0-dev/5a61bedef` (2021-10-21 10:32:45)
- `github.com/adlrocha/ipfs-lite` (2021-10-22 21:00:10)
- `go-ipfs/0.11.0-dev/23442df` (2021-10-23 11:03:07)

### Protocols

Newly discovered protocols:

- `/lilu.red/op/1/file` (2021-10-18 13:04:45)
- `/lilu.red/op/1/text` (2021-10-18 13:04:45)

### Classification

![](./plots-42/nodes.png)

Node classification:

- `offline` - A peer that was never seen online during the measurement period (always offline) but found in the DHT
- `dangling` - A peer that was seen going offline and online multiple times during the measurement period
- `oneoff` - A peer that was seen coming online and then going offline **only once** during the measurement period
- `online` - A peer that was not seen offline at all during the measurement period (always online)
- `left` - A peer that was online at the beginning of the measurement period, did go offline and didn't come back online
- `entered` - A peer that was offline at the beginning of the measurement period but appeared within and didn't go offline since then

### Top 10 Rotating Hosts

| IP-Address    | Country | Unique Peer IDs | Agent Versions |
|:------------- |:------- | ---------------:|:-------------- |
| 165.227.24.133 | US | 5233 | ['hydra-booster/0.7.4', None] |
| 82.165.18.239 | DE | 1609 | ['go-ipfs/0.4.22/', None] |
| 159.65.71.229 | US | 348 | [None] |
| 159.65.110.234 | US | 273 | [None] |
| 138.197.207.75 | US | 265 | [None] |
| 159.65.108.245 | US | 249 | [None] |
| 138.68.47.189 | US | 183 | [None] |
| 138.68.45.10 | US | 180 | [None] |
| 116.202.229.43 | DE | 148 | ['hydra-booster/0.7.4', None] |
| 123.157.156.218 | CN | 101 | [None] |

### Crawl Time Series

![](./plots-42/crawl-overview.png)

#### By Agent Version (selection)

![](./plots-42/crawl-properties.png)

## Churn

![](./plots-42/crawl-churn.png)

## Inter Arrival Time

![](./plots-42/cdf-inter-arrival-dangling.png)

## Agent Version Analysis

### Overall

![](./plots-42/agents-all.png)

Includes all peers that the crawler was able to connect to at least once (`dangling`, `online`, `oneoff`, `entered`)

### Dangling Nodes Only

![](./plots-42/agents-dangling.png)

Includes all peers that were seen going offline and online multiple times during the measurement.

### Online Nodes Only

![](./plots-42/agents-online.png)

Includes all peers that were not seen offline at all during the measurement period (always online).

### Oneoff Nodes Only

![](./plots-42/agents-oneoff.png)

Includes all peers that were seen coming online and then going offline **only once** during the measurement period

### Entered Nodes Only

![](./plots-42/agents-entered.png)

Includes all peers that were offline at the beginning of the measurement period but appeared within and didn't go offline since then.

## Geo location

### Resolution Statistics

![](./plots-42/geo-resolution.png)

Resolution Classification:

- `resolved` - The number of peer IDs that could be resolved to at least one IP address (excludes peers that are only reachable via circuit-relays)
- `unresolved` - The number of peer IDs that could not or just were not yet resolved to at least one IP address
- `no public ip` - The number of peer IDs that were found in the DHT but didn't have a public IP address
- `relay` - The number of peer IDs that were only reachable by circuit relays

### Unique IP Addresses

![](./plots-42/geo-unique-ip.png)

### Classification

![](./plots-42/geo-node-classification.png)

### Agents

![](./plots-42/geo-agents.png)


## Latencies

### Overall

![](./plots-42/latencies.png)

`Connect` measures the time it takes for the `libp2p` `host.Connect` call to return.

`Connect plus Crawl` includes the time of dialing, connecting and crawling the peer. `Crawling` means the time it takes for the FIND_NODE RPCs to resolve. Nebula is sending 15 of those with increasing common prefix lengths (CPLs) to the remote peer in parallel. 

### By Continent

![](./plots-42/geo-dial-latency-distribution.png)

## Cloud

The number next to `Total` indicates the number of unique IP addresses that went into this calculation.

### All

![](./plots-42/cloud-all.png)

### Classification

#### Offline

![](./plots-42/cloud-offline.png)

#### Online

![](./plots-42/cloud-online.png)

#### Dangling

![](./plots-42/cloud-dangling.png)

### Agent Version

#### Hydra

![](./plots-42/cloud-hydra.png)

#### ioi

![](./plots-42/cloud-ioi.png)

#### storm

![](./plots-42/cloud-storm.png)


## Terminology

- `visit` - Visiting a peer means dialing or connecting to it. Every time the crawler or monitoring task tries to dial or connect to a peer the following data is saved:
    ```sql
    id               SERIAL
    peer_id          SERIAL      NOT NULL -- this is now the internal database ID (not the peerID)
    crawl_id         INT                  -- can be null if this peer was visited from the monitoring task
    session_id       INT                  
    dial_duration    INTERVAL             -- The time it took to dial the peer or until an error occurred (NULL for crawl visits)
    connect_duration INTERVAL             -- The time it took to connect with the peer or until an error occurred (NULL for monitoring visits)
    crawl_duration   INTERVAL             -- The time it took to crawl the peer also if an error occurred (NULL for monitoring visits)
    updated_at       TIMESTAMPTZ NOT NULL 
    created_at       TIMESTAMPTZ NOT NULL 
    type             visit_type  NOT NULL -- either `dial` or `crawl`
    error            dial_error
    protocols_set_id INT                  -- a foreign key to the protocol set that this peer supported at this visit (NULL for monitoring visits as peers are just dialed)
    agent_version_id INT                  -- a foreign key to the peers agent version at this visit (NULL for monitoring visits as peers are just dialed)
    multi_addresses_set_id INT            -- a foreign key to the multi address set that was used to connect/dial for this visit
    ```

### Node classification:

- `offline` - A peer that was never seen online during the measurement period (always offline) but found in the DHT
- `dangling` - A peer that was seen going offline and online multiple times during the measurement period
- `oneoff` - A peer that was seen coming online and then going offline only once during the measurement period multiple times
- `online` - A peer that was not seen offline at all during the measurement period (always online)
- `left` - A peer that was online at the beginning of the measurement period, did go offline and didn't come back online
- `entered` - A peer that was offline at the beginning of the measurement period but appeared within and didn't go offline since then

### IP Resolution Classification:

- `resolved` - The number of peer IDs that could be resolved to at least one IP address (excludes peers that are only reachable by circuit-relays)
- `unresolved` - The number of peer IDs that could not or just were not yet resolved to at least one IP address
- `no public ip` - The number of peer IDs that were found in the DHT but didn't have a public IP address
- `relay` - The number of peer IDs that were only reachable by circuit relays

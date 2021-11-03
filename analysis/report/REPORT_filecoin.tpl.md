# Nebula Measurement Results Calendar Week {{ calendar_week }}

## General Information

The measurements were conducted on the following machine:

- `vCPU` - `4`
- `RAM` - `8GB`
- `Disk` - `160GB`
- `Datacenter` - `nbg1-dc3`
- `Country` - `Germany`
- `City` - `Nuremberg`

The following results show measurement data that was collected in calendar week {{ calendar_week }} from {{ measurement_start }} to {{ measurement_end }} in {{ year }}.

- Number of crawls `{{ crawl_count }}`
- Number of visits `{{ visit_count }}` ([what is a visit?](#terminology))
- Number of unique peer IDs visited `{{ peer_id_count }}`
- Number of unique IP addresses found `{{ ip_address_count }}`

Timestamps are in UTC if not mentioned otherwise.

### Protocols

Newly discovered protocols:
{% for p in new_protocols %}
- {{ p }}{% endfor %}

### Classification

![](./plots-{{ calendar_week }}/node_classifications.png)

Node classification:

- `offline` - A peer that was never seen online during the measurement period (always offline) but found in the DHT
- `dangling` - A peer that was seen going offline and online multiple times during the measurement period
- `oneoff` - A peer that was seen coming online and then going offline **only once** during the measurement period
- `online` - A peer that was not seen offline at all during the measurement period (always online)
- `left` - A peer that was online at the beginning of the measurement period, did go offline and didn't come back online
- `entered` - A peer that was offline at the beginning of the measurement period but appeared within and didn't go offline since then

### Crawl Time Series

![](./plots-{{ calendar_week }}/crawl-overview.png)

#### By Agent Version (selection)

![](./plots-{{ calendar_week }}/crawl-properties.png)

## Churn

![](./plots-{{ calendar_week }}/crawl-churn.png)

## Inter Arrival Time

![](./plots-{{ calendar_week }}/cdf-inter-arrival-dangling.png)

## Agent Version Analysis

### Overall

![](./plots-{{ calendar_week }}/agents-all.png)

Includes all peers that the crawler was able to connect to at least once (`dangling`, `online`, `oneoff`, `entered`)

### Dangling Nodes Only

![](./plots-{{ calendar_week }}/agents-dangling.png)

Includes all peers that were seen going offline and online multiple times during the measurement.

### Online Nodes Only

![](./plots-{{ calendar_week }}/agents-online.png)

Includes all peers that were not seen offline at all during the measurement period (always online).

### Oneoff Nodes Only

![](./plots-{{ calendar_week }}/agents-oneoff.png)

Includes all peers that were seen coming online and then going offline **only once** during the measurement period

### Entered Nodes Only

![](./plots-{{ calendar_week }}/agents-entered.png)

Includes all peers that were offline at the beginning of the measurement period but appeared within and didn't go offline since then.

## Geo location

### Resolution Statistics

![](./plots-{{ calendar_week }}/geo-resolution.png)

Resolution Classification:

- `resolved` - The number of peer IDs that could be resolved to at least one IP address (excludes peers that are only reachable via circuit-relays)
- `unresolved` - The number of peer IDs that could not or just were not yet resolved to at least one IP address
- `no public ip` - The number of peer IDs that were found in the DHT but didn't have a public IP address
- `relay` - The number of peer IDs that were only reachable by circuit relays

### Unique IP Addresses

![](./plots-{{ calendar_week }}/geo-unique-ip.png)

### Classification

![](./plots-{{ calendar_week }}/geo-node-classification.png)

### Agents

![](./plots-{{ calendar_week }}/geo-agents.png)


## Latencies

### Overall

![](./plots-{{ calendar_week }}/latencies.png)

`Connect` measures the time it takes for the `libp2p` `host.Connect` call to return.

`Connect plus Crawl` includes the time of dialing, connecting and crawling the peer. `Crawling` means the time it takes for the FIND_NODE RPCs to resolve. Nebula is sending 15 of those with increasing common prefix lengths (CPLs) to the remote peer in parallel. 

### By Continent

![](./plots-{{ calendar_week }}/latencies-geo.png)

## Terminology

- `visit` - Visiting a peer means dialing or connecting to it. Every time the crawler or monitoring task tries to dial or connect to a peer we consider this as _visiting_ it. Regardless of errors that may occur. 

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

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

- Number of crawls {{ crawl_count }}
- Number of visits {{ visit_count }}
- Number of unique peer IDs visited {{ peer_id_count }}

### Agent Versions

Newly discovered agent versions:
{% for av in new_agent_versions %}
- {{ av }}{% endfor %}

### Protocols

Newly discovered protocols:
{% for p in new_protocols %}
- {{ p }}{% endfor %}

### Classification

![](./plots-{{ calendar_week }}/nodes.png)

Node classification:

- `dangling` - A peer that was seen going offline and online during the measurement period (potentially multiple times)
- `offline` - A peer that was not seen online but found in the DHT during the measurement period (always offline)
- `online` - A peer that was not seen offline at all during the measurement period (always online)
- `left` - A peer that was online at the beginning of the measurement period, did go offline and didn't come back online
- `entered` - A peer that was offline at the beginning of the measurement period but appeared within and didn't go offline since then


### Crawl Time Series

![](./plots-{{ calendar_week }}/crawl-overview.png)

The top graph shows the number of dialable and undialable peers for each individual crawl. Further it shows the sum of both as `Total`.

The bottom graph shows the percentage of dialable peers in each crawl (`Dialable` / `Total`)

#### By Agent Version (selection)

![](./plots-{{ calendar_week }}/crawl-properties.png)

## Churn

![](./plots-{{ calendar_week }}/crawl-churn.png)

## Inter Arrival Time

![](./plots-{{ calendar_week }}/cdf-inter-arrival-dangling.png)

## Agent Version Analysis

### Overall

![](./plots-{{ calendar_week }}/agents-all.png)

These graphs show the agent version distribution that was observed during crawling the network. The number next to `Total` indicates the number of successful `crawl` visits that contribute to the distribution. 

### Dangling Nodes Only

![](./plots-{{ calendar_week }}/agents-dangling.png)

These graphs show the agent version distribution that was observed during crawling the network of only the dangling nodes. The number next to `Total` indicates the number of successful `crawl` visits that contribute to the distribution. 

### Online Nodes Only

![](./plots-{{ calendar_week }}/agents-online.png)

These graphs show the agent version distribution that was observed during crawling the network of the nodes that were online the whole measurement period (very stable peers). The number next to `Total` indicates the number of successful `crawl` visits that contribute to the distribution. 

## Geo location

### All

![](./plots-{{ calendar_week }}/geo-all.png)

Geo locations of all visited peers.

### Unique

![](./plots-{{ calendar_week }}/geo-unique.png)

This graph shows the country distribution of all seen unique IP addresses during the measurement period.

### Classification

#### Online

![](./plots-{{ calendar_week }}/geo-online.png)

#### Offline

![](./plots-{{ calendar_week }}/geo-offline.png)

#### Dangling

![](./plots-{{ calendar_week }}/geo-dangling.png)


### Agent Version

#### Hydra

![](./plots-{{ calendar_week }}/geo-hydra.png)

#### ioi

![](./plots-{{ calendar_week }}/geo-ioi.png)

#### storm

![](./plots-{{ calendar_week }}/geo-storm.png)

## Cloud

The number next to `Total` indicates the number of unique IP addresses that went into this calculation.

### All

![](./plots-{{ calendar_week }}/cloud-all.png)

### Classification

#### Offline

![](./plots-{{ calendar_week }}/cloud-offline.png)

#### Online

![](./plots-{{ calendar_week }}/cloud-online.png)

#### Dangling

![](./plots-{{ calendar_week }}/cloud-dangling.png)

### Agent Version

#### Hydra

![](./plots-{{ calendar_week }}/cloud-hydra.png)

#### ioi

![](./plots-{{ calendar_week }}/cloud-ioi.png)

#### storm

![](./plots-{{ calendar_week }}/cloud-storm.png)


## Latencies

### Overall

![](./plots-{{ calendar_week }}/latencies.png)

`Connect` measures the time it takes for the `libp2p` `host.Connect` call to return. This involves several hand shakes under the hood (includes the dial duration as well).

`Connect plus Crawl` includes the time of dialing, connecting (as explained above) and crawling the peer. `Crawling` means the time it takes for the FIND_NODE RPCs to resolve. Nebula is sending 15 of those with increasing common prefix lengths (CPLs) to the remote peer in parallel. 

### By Continent

![](./plots-{{ calendar_week }}/geo-dial-latency-distribution.png)


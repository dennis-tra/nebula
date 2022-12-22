# Nebula Measurement Results Calendar Week {{ calendar_week }} - {{ year }}

## Table of Contents

- [General Information](#general-information)
  - [Agent Versions](#agent-versions)
  - [Protocols](#protocols)
  - [Classification](#classification)
  - [Crawls](#crawls)
    - [Overall](#overall)
    - [By Agent Version](#by-agent-version)

## General Information

The following results show measurement data that were collected in calendar week {{ calendar_week }} in {{ year }} from `{{ measurement_start }}` to `{{ measurement_end }}`.

- Number of crawls `{{ crawl_count }}`
- Number of visits `{{ visit_count }}`
- Number of unique peer IDs visited `{{ peer_id_count }}`
- Number of unique IP addresses found `{{ ip_address_count }}`

Timestamps are in UTC if not mentioned otherwise.

### Agent Versions

Newly discovered agent versions:

{% for _, row in new_agent_versions.iterrows() %}
- `{{ row["agent_version"] }}` ({{ row["created_at"].strftime("%Y-%m-%d %H:%M:%S") }}){% endfor %}

Agent versions that were found to support at least one [storm specific protocol](#storm-specific-protocol):

{% for av in storm_agent_versions %}
- `{{ av }}`{% endfor %}

### Protocols

Newly discovered protocols:

{% for _, row in new_protocols.iterrows() %}
- `{{ row["protocol"] }}` ({{ row["created_at"].strftime("%Y-%m-%d %H:%M:%S") }}){% endfor %}

### Classification

![](./plots-{{ calendar_week }}/peer-classifications.png)

Peer classification:

| Classification | Description |
| --- | --- |
| `offline` | A peer that was never seen online during the measurement period (always offline) but found in the DHT |
| `dangling` | A peer that was seen going offline and online multiple times during the measurement period |
| `oneoff` | A peer that was seen coming online and then going offline **only once** during the measurement period |
| `online` | A peer that was not seen offline at all during the measurement period (always online) |
| `left` | A peer that was online at the beginning of the measurement period, did go offline and didn't come back online |
| `entered` | A peer that was offline at the beginning of the measurement period but appeared within and didn't go offline since then |

### Top 10 Rotating Nodes

| IP-Address    | Country | Unique Peer IDs | Agent Versions | Datacenter IP |
|:------------- |:------- | ---------------:|:-------------- | ------------- |{% for _, trn in top_rotating_nodes.iterrows() %}
| `{{ trn["ip_address"] }}` | {{ trn["country"] }} | {{ trn["peer_id_count"] }} | {{ trn["agent_versions"] }}| {{ trn["is_datacenter_ip"] }}  |{% endfor %}

> A "rotating nodes" is a node (as identified by a single IP address) that was found with multiple peer IDs.

### Crawls

#### Overall

![](./plots-{{ calendar_week }}/crawl-overview.png)

#### By Agent Version

![](./plots-{{ calendar_week }}/crawl-properties.png)

> Only the top 10 kubo versions appear in the right graph (due to lack of colors). The `0.8.x` versions do not contain disguised storm peers.

> `storm*` are `go-ipfs/0.8.0/48f94e2` peers that support at least one [storm specific protocol](#storm-specific-protocol).

## Churn

![](./plots-{{ calendar_week }}/peer-churn.png)

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

## Cloud

The number next to `Total` indicates the number of unique IP addresses that went into this calculation.

### All

![](./plots-{{ calendar_week }}/cloud-all.png)

### Classification

![](./plots-{{ calendar_week }}/cloud-classification.png)

### Agents

![](./plots-{{ calendar_week }}/cloud-agents.png)


## Top Updating Peers

| Peer ID           | Final AV     | # Transitions | Distinct AVs | # Distinct AVs |
|:----------------- |:------------ | ------------- |:------------ | -------------- |{% for tuh in top_updating_nodes %}
| `{{ tuh[1][:16] }}...` | `{{ tuh[2] }}` | {{ tuh[3] }}  | {{ '<br/>'.join(tuh[4]) }} | {{ tuh[5] }} |{% endfor %}

> `AV` = `Agent Version`
 
- `# Final AV` - The last agent version that was observed in this measurement period
- `# Transitions` - How often did this particular peer change its agent version
- `Distinct AVs` - Distinct agent version that this peer transitioned between (could have had a single AV multiple times)

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

### Cloud Providers

- `AWS` - Amazon Web Services
- `GCP` - Google Cloud Platform
- `Azure` - Microsoft Azure
- `DO` - Digital Ocean
- `OCI` - Oracle Cloud Infrastructure

### Storm Specific Protocols

The following protocol strings are unique for `storm` nodes according to [this Bitdefender paper](https://www.bitdefender.com/files/News/CaseStudies/study/376/Bitdefender-Whitepaper-IPStorm.pdf):

- `/sreque/*`
- `/shsk/*`
- `/sfst/*`
- `/sbst/*`
- `/sbpcp/*`
- `/sbptp/*`
- `/strelayp/*`
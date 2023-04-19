# Nebula Measurement Results Calendar Week {{ calendar_week }} - {{ year }}

## Table of Contents

- [General Information](#general-information)
  - [Agent Versions](#agent-versions)
  - [Protocols](#protocols)
  - [Classification](#classification)
  - [Top 10 Rotating Nodes](#top-10-rotating-nodes)
  - [Crawls](#crawls)
    - [Overall](#overall)
    - [By Agent Version](#by-agent-version)
- [Churn](#churn)
- [Inter Arrival Time](#inter-arrival-time)
- [Agent Version Analysis](#agent-version-analysis)
  - [Overall](#overall-1)
  - [Kubo](#kubo)
  - [Classification](#classification-1)
- [Geo location](#geo-location)
  - [Unique IP Addresses](#unique-ip-addresses)
  - [Classification](#classification-2)
  - [Agents](#agents)
- [Top Updating Peers](#top-updating-peers)
  - [Node classification:](#node-classification)
  - [IP Resolution Classification:](#ip-resolution-classification)
  - [Cloud Providers](#cloud-providers)
  - [Storm Specific Protocols](#storm-specific-protocols)

## General Information

The following results show measurement data that were collected in calendar week {{ calendar_week }} in {{ year }} from `{{ measurement_start }}` to `{{ measurement_end }}`.

- Number of crawls `{{ crawl_count }}`
- Number of visits `{{ visit_count }}`
  > Visiting a peer means dialing or connecting to it. Every time the crawler or monitoring process tries to dial or connect to a peer we consider this as _visiting_ it. Regardless of errors that may occur.
- Number of unique peer IDs visited `{{ visited_peer_id_count }}`
- Number of unique peer IDs discovered in the DHT `{{ discovered_peer_id_count }}`
- Number of unique IP addresses found `{{ ip_address_count }}`

Timestamps are in UTC if not mentioned otherwise.

### Agent Versions

Newly discovered agent versions:
{% for _, row in new_agent_versions.iterrows() %}
- `{{ row["agent_version"] }}` ({{ row["created_at"].strftime("%Y-%m-%d %H:%M:%S") }}){% endfor %}

Agent versions that were found to support at least one [storm specific protocol](#storm-specific-protocols):
{% for av in storm_agent_versions %}
- `{{ av }}`{% endfor %}

### Protocols

Newly discovered protocols:

{% for _, row in new_protocols.iterrows() %}
- `{{ row["protocol"] }}` ({{ row["created_at"].strftime("%Y-%m-%d %H:%M:%S") }}){% endfor %}

### Top 10 Rotating Nodes

A "rotating node" is a node (as identified by its IP address) that was found to host multiple peer IDs.

| IP-Address    | Country | Unique Peer IDs | Agent Versions | Datacenter IP |
|:------------- |:------- | ---------------:|:-------------- | ------------- |{% for _, trn in top_rotating_nodes.iterrows() %}
| `{{ trn["ip_address"] }}` | {{ trn["country"] }} | {{ trn["peer_id_count"] }} | {{ trn["agent_versions"] }}| {{ trn["is_datacenter_ip"] }}  |{% endfor %}

### Crawls

#### Overall

![Crawl Overview]({{ plots_dir }}/crawl-overview.png)

#### Classification

![Crawl Classifications]({{ plots_dir }}/crawl-classifications.png)

#### Agents

![Crawl Properties By Agent]({{ plots_dir }}/crawl-properties.png)

Only the top 10 kubo versions appear in the right graph (due to lack of colors) based on the average count in the time interval. The `0.8.x` versions **do not** contain disguised storm peers.

`storm*` are `{{ ", ".join(storm_star_agent_versions) }}` peers that support at least one [storm specific protocol](#storm-specific-protocols).

#### Errors

![Crawl Errors]({{ plots_dir }}/crawl-errors.png)

#### Total Peer IDs Discovered Classification

![Peer count by classification]({{ plots_dir }}/peer-classifications.png)

In the specified time interval from `{{ measurement_start }}` to `{{ measurement_end }}` we visited `{{ peer_id_count }}` unique peer IDs.
All peer IDs fall into one of the following classifications:

| Classification | Description |
| --- | --- |
| `offline` | A peer that was never seen online during the measurement period (always offline) but found in the DHT |
| `dangling` | A peer that was seen going offline and online multiple times during the measurement period |
| `oneoff` | A peer that was seen coming online and then going offline **only once** during the measurement period |
| `online` | A peer that was not seen offline at all during the measurement period (always online) |
| `left` | A peer that was online at the beginning of the measurement period, did go offline and didn't come back online |
| `entered` | A peer that was offline at the beginning of the measurement period but appeared within and didn't go offline since then |

#### Protocols

![Crawl Properties By Protocols]({{ plots_dir }}/crawl-protocols.png)

## Churn

![Peer Churn]({{ plots_dir }}/peer-churn.png)

Only the top 10 kubo versions appear in the right graph (due to lack of colors) based on the average count in the time interval. The `0.8.x` versions **do not** contain disguised storm peers. This graph also excludes peers that were online the whole time. You can read this graph as: if I see a peer joining the network, what's the likelihood for it to stay `X` hours in the network.

`storm*` are `{{ ", ".join(storm_star_agent_versions) }}` peers that support at least one [storm specific protocol](#storm-specific-protocols).

## Inter Arrival Time

![Inter Arrival Time]({{ plots_dir }}/peer-inter-arrival-time.png)

Only the top 10 kubo versions appear in the right graph (due to lack of colors) based on the average count in the time interval. The `0.8.x` versions **do not** contain disguised storm peers.

`storm*` are `{{ ", ".join(storm_star_agent_versions) }}` peers that support at least one [storm specific protocol](#storm-specific-protocols).

## Agent Version Analysis

### Overall

![Overall Agent Distribution]({{ plots_dir }}/agents-overall.png)

Includes all peers that the crawler was able to connect to at least once: `dangling`, `online`, `oneoff`, `entered`. Hence, the total number of peers is lower as the graph excludes `offline` and `left` peers (see [classification](#peer-classification)).

### Kubo

![Kubo Agent Distribution]({{ plots_dir }}/agents-kubo.png)

`storm` shows the `go-ipfs/0.8.0/48f94e2` peers that support at least one [storm specific protocol](#storm-specific-protocols).

### Classification

![Agents by Classification]({{ plots_dir }}/agents-classification.png)

The classifications are documented [here](#peer-classification).
`storm*` are `{{ ", ".join(storm_star_agent_versions) }}` peers that support at least one [storm specific protocol](#storm-specific-protocols).

## Geolocation

### Unique IP Addresses

![Unique IP addresses]({{ plots_dir }}/geo-unique-ip.png)

This graph shows all IP addresses that we found from `{{ measurement_start }}` to `{{ measurement_end }}` in the DHT and their geolocation distribution by country.

### Classification

![Peer Geolocation By Classification]({{ plots_dir }}/geo-peer-classification.png)

The classifications are documented [here](#peer-classification). 
The number in parentheses in the graph titles show the number of unique peer IDs that went into the specific subgraph.

### Agents

![Peer Geolocation By Agent]({{ plots_dir }}/geo-peer-agents.png)

`storm*` are `{{ ", ".join(storm_star_agent_versions) }}` peers that support at least one [storm specific protocol](#storm-specific-protocols).

## Datacenters

### Overall

![Overall Datacenter Distribution]({{ plots_dir }}/cloud-overall.png)

This graph shows all IP addresses that we found from `{{ measurement_start }}` to `{{ measurement_end }}` in the DHT and their datacenter association.

### Classification

![Datacenter Distribution By Classification]({{ plots_dir }}/cloud-classification.png)

The classifications are documented [here](#peer-classification). Note that the x-axes are different.

### Agents

![Datacenter Distribution By Agent]({{ plots_dir }}/cloud-agents.png)

The number in parentheses in the graph titles show the number of unique peer IDs that went into the specific subgraph.

`storm*` are `{{ ", ".join(storm_star_agent_versions) }}` peers that support at least one [storm specific protocol](#storm-specific-protocols).

### Peer Classification

| Classification | Description |
| --- | --- |
| `offline` | A peer that was never seen online during the measurement period (always offline) but found in the DHT |
| `dangling` | A peer that was seen going offline and online multiple times during the measurement period |
| `oneoff` | A peer that was seen coming online and then going offline **only once** during the measurement period |
| `online` | A peer that was not seen offline at all during the measurement period (always online) |
| `left` | A peer that was online at the beginning of the measurement period, did go offline and didn't come back online |
| `entered` | A peer that was offline at the beginning of the measurement period but appeared within and didn't go offline since then |

### Storm Specific Protocols

The following protocol strings are unique for `storm` nodes according to [this Bitdefender paper](https://www.bitdefender.com/files/News/CaseStudies/study/376/Bitdefender-Whitepaper-IPStorm.pdf):

- `/sreque/*`
- `/shsk/*`
- `/sfst/*`
- `/sbst/*`
- `/sbpcp/*`
- `/sbptp/*`
- `/strelayp/*`

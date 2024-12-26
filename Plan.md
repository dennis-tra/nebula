# Integrating Bitcoin Network Support into Nebula Crawler

This document outlines the challenges of adding support for the Bitcoin network to the Nebula crawler and provides a detailed list of the tasks required to achieve this integration. Adding support for the Bitcoin network to the Nebula crawler involves several challenges due to differences in network protocols and message handling. By systematically implementing the tasks outlined below, these challenges can be overcome. The integration will enable Nebula to crawl the Bitcoin network (and bitcoin adjacent networks) effectively, providing valuable insights and extending its capabilities to support a wider range of networks.

## Challenges in Adding Bitcoin Support to Nebula

1. **Adapting to the Bitcoin Wire Protocol**

   The Bitcoin network uses a unique wire protocol that differs from the protocols currently supported by Nebula. Integrating the `btcsuite/btcd/wire` package requires Nebula to understand and communicate using Bitcoin's protocol, including handling specific message formats, handshake procedures, and network behaviors.

2. **Integrating with Nebula's Interfaces**

   Nebula requires crawlers to implement specific interfaces such as `PeerInfo`, `CrawlDriverConfig`, and `CrawlDriver`. Ensuring the Bitcoin crawler adheres to these interfaces while accommodating the Bitcoin network's nuances is essential for seamless integration.

3. **Handling Unknown Messages and Errors**

   Initial attempts to crawl the Bitcoin network resulted in errors like `Failed to read message: received unknown message`. Resolving these errors is crucial to ensure proper communication and data exchange with Bitcoin peers.

4. **Updating the Command-Line Interface (CLI)**

   Nebula's CLI must be updated to include options relevant to the Bitcoin network. This involves adding new flags, updating parsing logic, and ensuring seamless integration with the existing `crawl` sub-command.

## Implementation Tasks

### 1. Create Interfaces for the Bitcoin Crawler

Implement the necessary interfaces to integrate the Bitcoin crawler into Nebula.

- **`PeerInfo`**

  - Define a `PeerInfo` struct specific to the Bitcoin network, including fields like IP address, port, protocol version, services offered, user agent, and last seen timestamp.
  - Ensure this struct implements all methods required by Nebula's `PeerInfo` interface.

- **`CrawlDriverConfig`**

  - Create a configuration struct for the Bitcoin crawler, including parameters such as connection timeouts, maximum peers, network parameters (e.g., mainnet, testnet), and logging options.
  - Ensure compatibility with Nebula's configuration handling mechanisms.

- **`CrawlDriver`**

  - Implement the `CrawlDriver` interface for the Bitcoin network.
  - Provide methods for starting the crawl, handling peer connections, sending and receiving messages, and gracefully shutting down the crawler.

- `core/core.go`
  - Implement the `Handler` and `WorkResult` interface for the Bitcoin network.
  - Provide methods for starting the crawl, handling peer connections, sending and receiving messages, and gracefully shutting down the crawler.

### 2. Add Placeholder Support for the Bitcoin Network

Update Nebula's main crawler code to include support for the Bitcoin network.

- **Add a New CLI Option**

  - Introduce a new command-line flag (e.g., `--network bitcoin`) to select the Bitcoin network.
  - Update the CLI parsing logic to recognize and process this new option.
  - Update help texts and usage instructions to reflect the addition.

- **Handle Network Type and Initialize Crawler**

  - Modify the main execution flow to handle the Bitcoin network type.
  - Instantiate the Bitcoin `CrawlDriver` when the network type is set to Bitcoin.
  - Ensure the crawler engine initializes correctly and begins the crawling process.

### 3. Investigate and Resolve Bitcoin Crawl Errors

Analyze and fix the errors encountered during Bitcoin crawling.

- **Analyze Error Messages**

  - Review error logs indicating `Failed to read message: received unknown message`.
  - Identify unrecognized or improperly handled message types.
  - Determine whether errors stem from unsupported messages, incorrect parsing, or protocol mismatches.

- **Update Message Handling Logic**

  - Implement support for all relevant Bitcoin message types using `btcsuite/btcd/wire`.
  - Ensure correct serialization and deserialization of messages.
  - Update the message processing loop to accommodate new message types and protocols.

- **Verify Protocol Compliance**

  - Ensure the crawler adheres to Bitcoin network protocol specifications.
  - Implement proper handshake procedures, including the `version` and `verack` message exchanges.
  - Handle protocol version negotiations and feature support declarations accurately.

### 4. Integrate the `btcsuite/btcd/wire` Package

Leverage the existing Bitcoin protocol implementation in Go.

- **Incorporate the Package**

  - Add `btcsuite/btcd/wire` as a dependency in the project using Go modules.
  - Manage the dependency version to ensure compatibility with Nebula.

- **Utilize Message Types and Structures**

  - Use provided message types (e.g., `MsgVersion`, `MsgAddr`, `MsgPing`) for communication.
  - Employ serialization and deserialization methods for message exchange.

- **Handle Network Magic Numbers**

  - Configure the crawler to use correct network magic numbers for mainnet, testnet, or regtest.
  - Ensure messages are correctly identified and routed based on these magic numbers.

### 5. Implement the basic Bitcoin network crawler

Follow up on task 2 by crawling the bitcoin nodes and visiting the peers.
This is the main task of this project.

- **DNS**

  - Current implementation at [btc-crawl](https://github.com/shazow/btc-crawl) uses dns entries that are converted to seeds.
  - This implementation needs to be translated to the nebula format.

- **Peer crawl**

  - The current [btc-crawl](https://github.com/shazow/btc-crawl) implementation already performs the peer crawl.
  - Btc crawl uses ip and port pairs instead of multi-addresses. This needs to be translated.
  - The bitcoin driver needs to be written to perform this crawl action, including the handshake and connection.

- **Logging**

  - Add logging to include Bitcoin-specific events and errors and to ensure ease of development.

## Plan Timeline and Milestones

- **By 20.01: Draft Presentation**

  - Prepare a draft presentation outlining progress, challenges, and next steps.
  - Include demonstrations of initial crawling capabilities if possible.

- **By 27.01: Presentation**

  - Finalize the presentation incorporating feedback from the draft.
  - Showcase working components and discuss how challenges were overcome.

- **By 04.02: Final Presentation**

  - Present the fully integrated Bitcoin crawler within Nebula.
  - Highlight achievements, performance metrics, and any remaining work.

- **By 31.03: Final Code Submission**

  - Complete all coding tasks, testing, and documentation.
  - Submit the final codebase with all features implemented and thoroughly tested.

## Additional Tasks and Considerations

- **Protocol Nuances**

  - Account for protocol variations between different Bitcoin client implementations.
  - Handle optional features and extensions gracefully.

- **Peer Discovery Mechanisms**

  - Implement peer discovery using `getaddr` and `addr` messages.
  - Consider seeding the crawler with known peers or utilizing DNS seeds.

- **Security Practices**

  - Validate all incoming data to prevent injection attacks or buffer overflows.
  - Implement encryption if connecting to peers that support it.

- **Documentation**

  - Update Nebula's documentation to include the Bitcoin crawler.
  - Provide usage instructions, configuration options, and troubleshooting guides.

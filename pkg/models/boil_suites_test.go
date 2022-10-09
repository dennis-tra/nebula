// Code generated by SQLBoiler 4.13.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import "testing"

// This test suite runs each operation test in parallel.
// Example, if your database has 3 tables, the suite will run:
// table1, table2 and table3 Delete in parallel
// table1, table2 and table3 Insert in parallel, and so forth.
// It does NOT run each operation group in parallel.
// Separating the tests thusly grants avoidance of Postgres deadlocks.
func TestParent(t *testing.T) {
	t.Run("AgentVersions", testAgentVersions)
	t.Run("CrawlProperties", testCrawlProperties)
	t.Run("Crawls", testCrawls)
	t.Run("IPAddresses", testIPAddresses)
	t.Run("Latencies", testLatencies)
	t.Run("MultiAddresses", testMultiAddresses)
	t.Run("Neighbors", testNeighbors)
	t.Run("PeerLogs", testPeerLogs)
	t.Run("Peers", testPeers)
	t.Run("Protocols", testProtocols)
	t.Run("ProtocolsSets", testProtocolsSets)
	t.Run("Sessions", testSessions)
	t.Run("SessionsCloseds", testSessionsCloseds)
	t.Run("SessionsOpens", testSessionsOpens)
	t.Run("Visits", testVisits)
}

func TestDelete(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsDelete)
	t.Run("CrawlProperties", testCrawlPropertiesDelete)
	t.Run("Crawls", testCrawlsDelete)
	t.Run("IPAddresses", testIPAddressesDelete)
	t.Run("Latencies", testLatenciesDelete)
	t.Run("MultiAddresses", testMultiAddressesDelete)
	t.Run("Neighbors", testNeighborsDelete)
	t.Run("PeerLogs", testPeerLogsDelete)
	t.Run("Peers", testPeersDelete)
	t.Run("Protocols", testProtocolsDelete)
	t.Run("ProtocolsSets", testProtocolsSetsDelete)
	t.Run("Sessions", testSessionsDelete)
	t.Run("SessionsCloseds", testSessionsClosedsDelete)
	t.Run("SessionsOpens", testSessionsOpensDelete)
	t.Run("Visits", testVisitsDelete)
}

func TestQueryDeleteAll(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsQueryDeleteAll)
	t.Run("CrawlProperties", testCrawlPropertiesQueryDeleteAll)
	t.Run("Crawls", testCrawlsQueryDeleteAll)
	t.Run("IPAddresses", testIPAddressesQueryDeleteAll)
	t.Run("Latencies", testLatenciesQueryDeleteAll)
	t.Run("MultiAddresses", testMultiAddressesQueryDeleteAll)
	t.Run("Neighbors", testNeighborsQueryDeleteAll)
	t.Run("PeerLogs", testPeerLogsQueryDeleteAll)
	t.Run("Peers", testPeersQueryDeleteAll)
	t.Run("Protocols", testProtocolsQueryDeleteAll)
	t.Run("ProtocolsSets", testProtocolsSetsQueryDeleteAll)
	t.Run("Sessions", testSessionsQueryDeleteAll)
	t.Run("SessionsCloseds", testSessionsClosedsQueryDeleteAll)
	t.Run("SessionsOpens", testSessionsOpensQueryDeleteAll)
	t.Run("Visits", testVisitsQueryDeleteAll)
}

func TestSliceDeleteAll(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsSliceDeleteAll)
	t.Run("CrawlProperties", testCrawlPropertiesSliceDeleteAll)
	t.Run("Crawls", testCrawlsSliceDeleteAll)
	t.Run("IPAddresses", testIPAddressesSliceDeleteAll)
	t.Run("Latencies", testLatenciesSliceDeleteAll)
	t.Run("MultiAddresses", testMultiAddressesSliceDeleteAll)
	t.Run("Neighbors", testNeighborsSliceDeleteAll)
	t.Run("PeerLogs", testPeerLogsSliceDeleteAll)
	t.Run("Peers", testPeersSliceDeleteAll)
	t.Run("Protocols", testProtocolsSliceDeleteAll)
	t.Run("ProtocolsSets", testProtocolsSetsSliceDeleteAll)
	t.Run("Sessions", testSessionsSliceDeleteAll)
	t.Run("SessionsCloseds", testSessionsClosedsSliceDeleteAll)
	t.Run("SessionsOpens", testSessionsOpensSliceDeleteAll)
	t.Run("Visits", testVisitsSliceDeleteAll)
}

func TestExists(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsExists)
	t.Run("CrawlProperties", testCrawlPropertiesExists)
	t.Run("Crawls", testCrawlsExists)
	t.Run("IPAddresses", testIPAddressesExists)
	t.Run("Latencies", testLatenciesExists)
	t.Run("MultiAddresses", testMultiAddressesExists)
	t.Run("Neighbors", testNeighborsExists)
	t.Run("PeerLogs", testPeerLogsExists)
	t.Run("Peers", testPeersExists)
	t.Run("Protocols", testProtocolsExists)
	t.Run("ProtocolsSets", testProtocolsSetsExists)
	t.Run("Sessions", testSessionsExists)
	t.Run("SessionsCloseds", testSessionsClosedsExists)
	t.Run("SessionsOpens", testSessionsOpensExists)
	t.Run("Visits", testVisitsExists)
}

func TestFind(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsFind)
	t.Run("CrawlProperties", testCrawlPropertiesFind)
	t.Run("Crawls", testCrawlsFind)
	t.Run("IPAddresses", testIPAddressesFind)
	t.Run("Latencies", testLatenciesFind)
	t.Run("MultiAddresses", testMultiAddressesFind)
	t.Run("Neighbors", testNeighborsFind)
	t.Run("PeerLogs", testPeerLogsFind)
	t.Run("Peers", testPeersFind)
	t.Run("Protocols", testProtocolsFind)
	t.Run("ProtocolsSets", testProtocolsSetsFind)
	t.Run("Sessions", testSessionsFind)
	t.Run("SessionsCloseds", testSessionsClosedsFind)
	t.Run("SessionsOpens", testSessionsOpensFind)
	t.Run("Visits", testVisitsFind)
}

func TestBind(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsBind)
	t.Run("CrawlProperties", testCrawlPropertiesBind)
	t.Run("Crawls", testCrawlsBind)
	t.Run("IPAddresses", testIPAddressesBind)
	t.Run("Latencies", testLatenciesBind)
	t.Run("MultiAddresses", testMultiAddressesBind)
	t.Run("Neighbors", testNeighborsBind)
	t.Run("PeerLogs", testPeerLogsBind)
	t.Run("Peers", testPeersBind)
	t.Run("Protocols", testProtocolsBind)
	t.Run("ProtocolsSets", testProtocolsSetsBind)
	t.Run("Sessions", testSessionsBind)
	t.Run("SessionsCloseds", testSessionsClosedsBind)
	t.Run("SessionsOpens", testSessionsOpensBind)
	t.Run("Visits", testVisitsBind)
}

func TestOne(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsOne)
	t.Run("CrawlProperties", testCrawlPropertiesOne)
	t.Run("Crawls", testCrawlsOne)
	t.Run("IPAddresses", testIPAddressesOne)
	t.Run("Latencies", testLatenciesOne)
	t.Run("MultiAddresses", testMultiAddressesOne)
	t.Run("Neighbors", testNeighborsOne)
	t.Run("PeerLogs", testPeerLogsOne)
	t.Run("Peers", testPeersOne)
	t.Run("Protocols", testProtocolsOne)
	t.Run("ProtocolsSets", testProtocolsSetsOne)
	t.Run("Sessions", testSessionsOne)
	t.Run("SessionsCloseds", testSessionsClosedsOne)
	t.Run("SessionsOpens", testSessionsOpensOne)
	t.Run("Visits", testVisitsOne)
}

func TestAll(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsAll)
	t.Run("CrawlProperties", testCrawlPropertiesAll)
	t.Run("Crawls", testCrawlsAll)
	t.Run("IPAddresses", testIPAddressesAll)
	t.Run("Latencies", testLatenciesAll)
	t.Run("MultiAddresses", testMultiAddressesAll)
	t.Run("Neighbors", testNeighborsAll)
	t.Run("PeerLogs", testPeerLogsAll)
	t.Run("Peers", testPeersAll)
	t.Run("Protocols", testProtocolsAll)
	t.Run("ProtocolsSets", testProtocolsSetsAll)
	t.Run("Sessions", testSessionsAll)
	t.Run("SessionsCloseds", testSessionsClosedsAll)
	t.Run("SessionsOpens", testSessionsOpensAll)
	t.Run("Visits", testVisitsAll)
}

func TestCount(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsCount)
	t.Run("CrawlProperties", testCrawlPropertiesCount)
	t.Run("Crawls", testCrawlsCount)
	t.Run("IPAddresses", testIPAddressesCount)
	t.Run("Latencies", testLatenciesCount)
	t.Run("MultiAddresses", testMultiAddressesCount)
	t.Run("Neighbors", testNeighborsCount)
	t.Run("PeerLogs", testPeerLogsCount)
	t.Run("Peers", testPeersCount)
	t.Run("Protocols", testProtocolsCount)
	t.Run("ProtocolsSets", testProtocolsSetsCount)
	t.Run("Sessions", testSessionsCount)
	t.Run("SessionsCloseds", testSessionsClosedsCount)
	t.Run("SessionsOpens", testSessionsOpensCount)
	t.Run("Visits", testVisitsCount)
}

func TestHooks(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsHooks)
	t.Run("CrawlProperties", testCrawlPropertiesHooks)
	t.Run("Crawls", testCrawlsHooks)
	t.Run("IPAddresses", testIPAddressesHooks)
	t.Run("Latencies", testLatenciesHooks)
	t.Run("MultiAddresses", testMultiAddressesHooks)
	t.Run("Neighbors", testNeighborsHooks)
	t.Run("PeerLogs", testPeerLogsHooks)
	t.Run("Peers", testPeersHooks)
	t.Run("Protocols", testProtocolsHooks)
	t.Run("ProtocolsSets", testProtocolsSetsHooks)
	t.Run("Sessions", testSessionsHooks)
	t.Run("SessionsCloseds", testSessionsClosedsHooks)
	t.Run("SessionsOpens", testSessionsOpensHooks)
	t.Run("Visits", testVisitsHooks)
}

func TestInsert(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsInsert)
	t.Run("AgentVersions", testAgentVersionsInsertWhitelist)
	t.Run("CrawlProperties", testCrawlPropertiesInsert)
	t.Run("CrawlProperties", testCrawlPropertiesInsertWhitelist)
	t.Run("Crawls", testCrawlsInsert)
	t.Run("Crawls", testCrawlsInsertWhitelist)
	t.Run("IPAddresses", testIPAddressesInsert)
	t.Run("IPAddresses", testIPAddressesInsertWhitelist)
	t.Run("Latencies", testLatenciesInsert)
	t.Run("Latencies", testLatenciesInsertWhitelist)
	t.Run("MultiAddresses", testMultiAddressesInsert)
	t.Run("MultiAddresses", testMultiAddressesInsertWhitelist)
	t.Run("Neighbors", testNeighborsInsert)
	t.Run("Neighbors", testNeighborsInsertWhitelist)
	t.Run("PeerLogs", testPeerLogsInsert)
	t.Run("PeerLogs", testPeerLogsInsertWhitelist)
	t.Run("Peers", testPeersInsert)
	t.Run("Peers", testPeersInsertWhitelist)
	t.Run("Protocols", testProtocolsInsert)
	t.Run("Protocols", testProtocolsInsertWhitelist)
	t.Run("ProtocolsSets", testProtocolsSetsInsert)
	t.Run("ProtocolsSets", testProtocolsSetsInsertWhitelist)
	t.Run("Sessions", testSessionsInsert)
	t.Run("Sessions", testSessionsInsertWhitelist)
	t.Run("SessionsCloseds", testSessionsClosedsInsert)
	t.Run("SessionsCloseds", testSessionsClosedsInsertWhitelist)
	t.Run("SessionsOpens", testSessionsOpensInsert)
	t.Run("SessionsOpens", testSessionsOpensInsertWhitelist)
	t.Run("Visits", testVisitsInsert)
	t.Run("Visits", testVisitsInsertWhitelist)
}

// TestToOne tests cannot be run in parallel
// or deadlocks can occur.
func TestToOne(t *testing.T) {
	t.Run("CrawlPropertyToAgentVersionUsingAgentVersion", testCrawlPropertyToOneAgentVersionUsingAgentVersion)
	t.Run("CrawlPropertyToCrawlUsingCrawl", testCrawlPropertyToOneCrawlUsingCrawl)
	t.Run("CrawlPropertyToProtocolUsingProtocol", testCrawlPropertyToOneProtocolUsingProtocol)
	t.Run("IPAddressToMultiAddressUsingMultiAddress", testIPAddressToOneMultiAddressUsingMultiAddress)
	t.Run("LatencyToPeerUsingPeer", testLatencyToOnePeerUsingPeer)
	t.Run("NeighborToCrawlUsingCrawl", testNeighborToOneCrawlUsingCrawl)
	t.Run("NeighborToPeerUsingPeer", testNeighborToOnePeerUsingPeer)
	t.Run("PeerLogToPeerUsingPeer", testPeerLogToOnePeerUsingPeer)
	t.Run("PeerToAgentVersionUsingAgentVersion", testPeerToOneAgentVersionUsingAgentVersion)
	t.Run("PeerToProtocolsSetUsingProtocolsSet", testPeerToOneProtocolsSetUsingProtocolsSet)
	t.Run("SessionsOpenToPeerUsingPeer", testSessionsOpenToOnePeerUsingPeer)
}

// TestOneToOne tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOne(t *testing.T) {
	t.Run("PeerToSessionsOpenUsingSessionsOpen", testPeerOneToOneSessionsOpenUsingSessionsOpen)
}

// TestToMany tests cannot be run in parallel
// or deadlocks can occur.
func TestToMany(t *testing.T) {
	t.Run("AgentVersionToCrawlProperties", testAgentVersionToManyCrawlProperties)
	t.Run("AgentVersionToPeers", testAgentVersionToManyPeers)
	t.Run("CrawlToCrawlProperties", testCrawlToManyCrawlProperties)
	t.Run("CrawlToNeighbors", testCrawlToManyNeighbors)
	t.Run("MultiAddressToIPAddresses", testMultiAddressToManyIPAddresses)
	t.Run("MultiAddressToPeers", testMultiAddressToManyPeers)
	t.Run("PeerToLatencies", testPeerToManyLatencies)
	t.Run("PeerToNeighbors", testPeerToManyNeighbors)
	t.Run("PeerToPeerLogs", testPeerToManyPeerLogs)
	t.Run("PeerToMultiAddresses", testPeerToManyMultiAddresses)
	t.Run("ProtocolToCrawlProperties", testProtocolToManyCrawlProperties)
	t.Run("ProtocolsSetToPeers", testProtocolsSetToManyPeers)
}

// TestToOneSet tests cannot be run in parallel
// or deadlocks can occur.
func TestToOneSet(t *testing.T) {
	t.Run("CrawlPropertyToAgentVersionUsingCrawlProperties", testCrawlPropertyToOneSetOpAgentVersionUsingAgentVersion)
	t.Run("CrawlPropertyToCrawlUsingCrawlProperties", testCrawlPropertyToOneSetOpCrawlUsingCrawl)
	t.Run("CrawlPropertyToProtocolUsingCrawlProperties", testCrawlPropertyToOneSetOpProtocolUsingProtocol)
	t.Run("IPAddressToMultiAddressUsingIPAddresses", testIPAddressToOneSetOpMultiAddressUsingMultiAddress)
	t.Run("LatencyToPeerUsingLatencies", testLatencyToOneSetOpPeerUsingPeer)
	t.Run("NeighborToCrawlUsingNeighbors", testNeighborToOneSetOpCrawlUsingCrawl)
	t.Run("NeighborToPeerUsingNeighbors", testNeighborToOneSetOpPeerUsingPeer)
	t.Run("PeerLogToPeerUsingPeerLogs", testPeerLogToOneSetOpPeerUsingPeer)
	t.Run("PeerToAgentVersionUsingPeers", testPeerToOneSetOpAgentVersionUsingAgentVersion)
	t.Run("PeerToProtocolsSetUsingPeers", testPeerToOneSetOpProtocolsSetUsingProtocolsSet)
	t.Run("SessionsOpenToPeerUsingSessionsOpen", testSessionsOpenToOneSetOpPeerUsingPeer)
}

// TestToOneRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestToOneRemove(t *testing.T) {
	t.Run("CrawlPropertyToAgentVersionUsingCrawlProperties", testCrawlPropertyToOneRemoveOpAgentVersionUsingAgentVersion)
	t.Run("CrawlPropertyToProtocolUsingCrawlProperties", testCrawlPropertyToOneRemoveOpProtocolUsingProtocol)
	t.Run("PeerToAgentVersionUsingPeers", testPeerToOneRemoveOpAgentVersionUsingAgentVersion)
	t.Run("PeerToProtocolsSetUsingPeers", testPeerToOneRemoveOpProtocolsSetUsingProtocolsSet)
}

// TestOneToOneSet tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOneSet(t *testing.T) {
	t.Run("PeerToSessionsOpenUsingSessionsOpen", testPeerOneToOneSetOpSessionsOpenUsingSessionsOpen)
}

// TestOneToOneRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOneRemove(t *testing.T) {}

// TestToManyAdd tests cannot be run in parallel
// or deadlocks can occur.
func TestToManyAdd(t *testing.T) {
	t.Run("AgentVersionToCrawlProperties", testAgentVersionToManyAddOpCrawlProperties)
	t.Run("AgentVersionToPeers", testAgentVersionToManyAddOpPeers)
	t.Run("CrawlToCrawlProperties", testCrawlToManyAddOpCrawlProperties)
	t.Run("CrawlToNeighbors", testCrawlToManyAddOpNeighbors)
	t.Run("MultiAddressToIPAddresses", testMultiAddressToManyAddOpIPAddresses)
	t.Run("MultiAddressToPeers", testMultiAddressToManyAddOpPeers)
	t.Run("PeerToLatencies", testPeerToManyAddOpLatencies)
	t.Run("PeerToNeighbors", testPeerToManyAddOpNeighbors)
	t.Run("PeerToPeerLogs", testPeerToManyAddOpPeerLogs)
	t.Run("PeerToMultiAddresses", testPeerToManyAddOpMultiAddresses)
	t.Run("ProtocolToCrawlProperties", testProtocolToManyAddOpCrawlProperties)
	t.Run("ProtocolsSetToPeers", testProtocolsSetToManyAddOpPeers)
}

// TestToManySet tests cannot be run in parallel
// or deadlocks can occur.
func TestToManySet(t *testing.T) {
	t.Run("AgentVersionToCrawlProperties", testAgentVersionToManySetOpCrawlProperties)
	t.Run("AgentVersionToPeers", testAgentVersionToManySetOpPeers)
	t.Run("MultiAddressToPeers", testMultiAddressToManySetOpPeers)
	t.Run("PeerToMultiAddresses", testPeerToManySetOpMultiAddresses)
	t.Run("ProtocolToCrawlProperties", testProtocolToManySetOpCrawlProperties)
	t.Run("ProtocolsSetToPeers", testProtocolsSetToManySetOpPeers)
}

// TestToManyRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestToManyRemove(t *testing.T) {
	t.Run("AgentVersionToCrawlProperties", testAgentVersionToManyRemoveOpCrawlProperties)
	t.Run("AgentVersionToPeers", testAgentVersionToManyRemoveOpPeers)
	t.Run("MultiAddressToPeers", testMultiAddressToManyRemoveOpPeers)
	t.Run("PeerToMultiAddresses", testPeerToManyRemoveOpMultiAddresses)
	t.Run("ProtocolToCrawlProperties", testProtocolToManyRemoveOpCrawlProperties)
	t.Run("ProtocolsSetToPeers", testProtocolsSetToManyRemoveOpPeers)
}

func TestReload(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsReload)
	t.Run("CrawlProperties", testCrawlPropertiesReload)
	t.Run("Crawls", testCrawlsReload)
	t.Run("IPAddresses", testIPAddressesReload)
	t.Run("Latencies", testLatenciesReload)
	t.Run("MultiAddresses", testMultiAddressesReload)
	t.Run("Neighbors", testNeighborsReload)
	t.Run("PeerLogs", testPeerLogsReload)
	t.Run("Peers", testPeersReload)
	t.Run("Protocols", testProtocolsReload)
	t.Run("ProtocolsSets", testProtocolsSetsReload)
	t.Run("Sessions", testSessionsReload)
	t.Run("SessionsCloseds", testSessionsClosedsReload)
	t.Run("SessionsOpens", testSessionsOpensReload)
	t.Run("Visits", testVisitsReload)
}

func TestReloadAll(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsReloadAll)
	t.Run("CrawlProperties", testCrawlPropertiesReloadAll)
	t.Run("Crawls", testCrawlsReloadAll)
	t.Run("IPAddresses", testIPAddressesReloadAll)
	t.Run("Latencies", testLatenciesReloadAll)
	t.Run("MultiAddresses", testMultiAddressesReloadAll)
	t.Run("Neighbors", testNeighborsReloadAll)
	t.Run("PeerLogs", testPeerLogsReloadAll)
	t.Run("Peers", testPeersReloadAll)
	t.Run("Protocols", testProtocolsReloadAll)
	t.Run("ProtocolsSets", testProtocolsSetsReloadAll)
	t.Run("Sessions", testSessionsReloadAll)
	t.Run("SessionsCloseds", testSessionsClosedsReloadAll)
	t.Run("SessionsOpens", testSessionsOpensReloadAll)
	t.Run("Visits", testVisitsReloadAll)
}

func TestSelect(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsSelect)
	t.Run("CrawlProperties", testCrawlPropertiesSelect)
	t.Run("Crawls", testCrawlsSelect)
	t.Run("IPAddresses", testIPAddressesSelect)
	t.Run("Latencies", testLatenciesSelect)
	t.Run("MultiAddresses", testMultiAddressesSelect)
	t.Run("Neighbors", testNeighborsSelect)
	t.Run("PeerLogs", testPeerLogsSelect)
	t.Run("Peers", testPeersSelect)
	t.Run("Protocols", testProtocolsSelect)
	t.Run("ProtocolsSets", testProtocolsSetsSelect)
	t.Run("Sessions", testSessionsSelect)
	t.Run("SessionsCloseds", testSessionsClosedsSelect)
	t.Run("SessionsOpens", testSessionsOpensSelect)
	t.Run("Visits", testVisitsSelect)
}

func TestUpdate(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsUpdate)
	t.Run("CrawlProperties", testCrawlPropertiesUpdate)
	t.Run("Crawls", testCrawlsUpdate)
	t.Run("IPAddresses", testIPAddressesUpdate)
	t.Run("Latencies", testLatenciesUpdate)
	t.Run("MultiAddresses", testMultiAddressesUpdate)
	t.Run("Neighbors", testNeighborsUpdate)
	t.Run("PeerLogs", testPeerLogsUpdate)
	t.Run("Peers", testPeersUpdate)
	t.Run("Protocols", testProtocolsUpdate)
	t.Run("ProtocolsSets", testProtocolsSetsUpdate)
	t.Run("Sessions", testSessionsUpdate)
	t.Run("SessionsCloseds", testSessionsClosedsUpdate)
	t.Run("SessionsOpens", testSessionsOpensUpdate)
	t.Run("Visits", testVisitsUpdate)
}

func TestSliceUpdateAll(t *testing.T) {
	t.Run("AgentVersions", testAgentVersionsSliceUpdateAll)
	t.Run("CrawlProperties", testCrawlPropertiesSliceUpdateAll)
	t.Run("Crawls", testCrawlsSliceUpdateAll)
	t.Run("IPAddresses", testIPAddressesSliceUpdateAll)
	t.Run("Latencies", testLatenciesSliceUpdateAll)
	t.Run("MultiAddresses", testMultiAddressesSliceUpdateAll)
	t.Run("Neighbors", testNeighborsSliceUpdateAll)
	t.Run("PeerLogs", testPeerLogsSliceUpdateAll)
	t.Run("Peers", testPeersSliceUpdateAll)
	t.Run("Protocols", testProtocolsSliceUpdateAll)
	t.Run("ProtocolsSets", testProtocolsSetsSliceUpdateAll)
	t.Run("Sessions", testSessionsSliceUpdateAll)
	t.Run("SessionsCloseds", testSessionsClosedsSliceUpdateAll)
	t.Run("SessionsOpens", testSessionsOpensSliceUpdateAll)
	t.Run("Visits", testVisitsSliceUpdateAll)
}

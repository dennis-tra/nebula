# discv4

## Behaviour Investigation

### `enode://e386566f35e376d6d9ae91678ac7c788fa0e558b449f66213207365e4c9b6173dd65b948a1f1e5b1178a7b04fe4d2f43b88054d9763405f994187cabdf39e6fd@44.213.54.129:30308`

- Always returns only three peers
- Returns different peers for the same key
- Returns more than 16 peers for the same key

### `enode://d860a01f9722d78051619d1e2351aba3f43f943f6f00718d1b9baa4101932a1f5011f16bb2b1bb35db20d6fe28fa0bf09636d26a87d31de9ec6203eeedb1f666@18.138.108.67:30303`

- Returns 16 peers to query
- Returns the same 16 peers for the same key


###

- some nodes 

### Implementation Investigation


peer: `enode://d860a01f9722d78051619d1e2351aba3f43f943f6f00718d1b9baa4101932a1f5011f16bb2b1bb35db20d6fe28fa0bf09636d26a87d31de9ec6203eeedb1f666@18.138.108.67:30303`

- Crawling sequentially yields 203 unique peers in the RT
- Crawling concurrently yields 128/115 unique peers in the RT

This is because if we're sending FIND_NODE RPCs concurrently to the remote peer
the implementation registers so-called "matchers" that will match the response
to the request. However, because the whole protocol uses raw UDP packets and the
requests don't have any ID associated with it, the response cannot be matched
properly. The implementation forwards the response to ALL matchers. This means
If we do FIND_NODE RPCs to the first 16 buckets, we register matchers for all
of them. The responses come back and are forwarded to all matchers. After the
matchers have seen enough peers, they get deregistered. This means many matchers
might see the same peers and get deregistered too early or get assigned the
wrong peers.

My solution here is to not forward a response from a node to all matchers but
only to a single one even if it's not the correct one associated with the
request. After all, we are only interested in the contents of the entire routing
table. This means it's not really an issue if the request for bucket 2 will
return the peers of bucket 10, for example. In the end, we throw all the results
together anyway.

-> Crawling concurrently and forwarding responses to only a single matcher -> 204 peers
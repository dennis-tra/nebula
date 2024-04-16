package config

var (

	// BootstrapPeersFilecoin extracted from:
	// https://github.com/filecoin-project/lotus/blob/b691adc4874e5e28353f036c077c08ef00ec3b2b/build/bootstrap/mainnet.pi
	BootstrapPeersFilecoin = []string{
		"/dns4/lotus-bootstrap.ipfsforce.com/tcp/41778/p2p/12D3KooWGhufNmZHF3sv48aQeS13ng5XVJZ9E6qy2Ms4VzqeUsHk",
		"/dns4/bootstrap-0.starpool.in/tcp/12757/p2p/12D3KooWGHpBMeZbestVEWkfdnC9u7p6uFHXL1n7m1ZBqsEmiUzz",
		"/dns4/bootstrap-1.starpool.in/tcp/12757/p2p/12D3KooWQZrGH1PxSNZPum99M1zNvjNFM33d1AAu5DcvdHptuU7u",
		"/dns4/node.glif.io/tcp/1235/p2p/12D3KooWBF8cpp65hp2u9LK5mh19x67ftAam84z9LsfaquTDSBpt",
		"/dns4/bootstarp-0.1475.io/tcp/61256/p2p/12D3KooWRzCVDwHUkgdK7eRgnoXbjDAELhxPErjHzbRLguSV1aRt",
		"/dns4/bootstrap-venus.mainnet.filincubator.com/tcp/8888/p2p/QmQu8C6deXwKvJP2D8B6QGyhngc3ZiDnFzEHBDx8yeBXST",
		"/dns4/bootstrap-mainnet-0.chainsafe-fil.io/tcp/34000/p2p/12D3KooWKKkCZbcigsWTEu1cgNetNbZJqeNtysRtFpq7DTqw3eqH",
		"/dns4/bootstrap-mainnet-1.chainsafe-fil.io/tcp/34000/p2p/12D3KooWGnkd9GQKo3apkShQDaq1d6cKJJmsVe6KiQkacUk1T8oZ",
		"/dns4/bootstrap-mainnet-2.chainsafe-fil.io/tcp/34000/p2p/12D3KooWHQRSDFv4FvAjtU32shQ7znz7oRbLBryXzZ9NMK2feyyH",
	}

	// BootstrapPeersKusama extracted from:
	//   https://github.com/paritytech/polkadot-sdk/blob/master/polkadot/node/service/chain-specs/kusama.json
	BootstrapPeersKusama = []string{
		"/dns/kusama-bootnode-0.polkadot.io/tcp/30333/p2p/12D3KooWSueCPH3puP2PcvqPJdNaDNF3jMZjtJtDiSy35pWrbt5h",
		"/dns/kusama-bootnode-0.polkadot.io/tcp/30334/ws/p2p/12D3KooWSueCPH3puP2PcvqPJdNaDNF3jMZjtJtDiSy35pWrbt5h",
		"/dns/kusama-bootnode-0.polkadot.io/tcp/443/wss/p2p/12D3KooWSueCPH3puP2PcvqPJdNaDNF3jMZjtJtDiSy35pWrbt5h",
		"/dns/kusama-bootnode-1.polkadot.io/tcp/30333/p2p/12D3KooWQKqane1SqWJNWMQkbia9qiMWXkcHtAdfW5eVF8hbwEDw",
		"/dns/kusama-bootnode-1.polkadot.io/tcp/30334/ws/p2p/12D3KooWQKqane1SqWJNWMQkbia9qiMWXkcHtAdfW5eVF8hbwEDw",
		"/dns/kusama-bootnode-1.polkadot.io/tcp/443/wss/p2p/12D3KooWQKqane1SqWJNWMQkbia9qiMWXkcHtAdfW5eVF8hbwEDw",
		"/dns/kusama-boot.dwellir.com/tcp/30333/ws/p2p/12D3KooWFj2ndawdYyk2spc42Y2arYwb2TUoHLHFAsKuHRzWXwoJ",
		"/dns/kusama-boot.dwellir.com/tcp/443/wss/p2p/12D3KooWFj2ndawdYyk2spc42Y2arYwb2TUoHLHFAsKuHRzWXwoJ",
		"/dns/boot.stake.plus/tcp/31333/p2p/12D3KooWLa1UyG5xLPds2GbiRBCTJjpsVwRWHWN7Dff14yiNJRpR",
		"/dns/boot.stake.plus/tcp/31334/wss/p2p/12D3KooWLa1UyG5xLPds2GbiRBCTJjpsVwRWHWN7Dff14yiNJRpR",
		"/dns/boot-node.helikon.io/tcp/7060/p2p/12D3KooWL4KPqfAsPE2aY1g5Zo1CxsDwcdJ7mmAghK7cg6M2fdbD",
		"/dns/boot-node.helikon.io/tcp/7062/wss/p2p/12D3KooWL4KPqfAsPE2aY1g5Zo1CxsDwcdJ7mmAghK7cg6M2fdbD",
		"/dns/kusama.bootnode.amforc.com/tcp/30333/p2p/12D3KooWLx6nsj6Fpd8biP1VDyuCUjazvRiGWyBam8PsqRJkbUb9",
		"/dns/kusama.bootnode.amforc.com/tcp/30334/wss/p2p/12D3KooWLx6nsj6Fpd8biP1VDyuCUjazvRiGWyBam8PsqRJkbUb9",
		"/dns/kusama.bootnodes.polkadotters.com/tcp/30311/p2p/12D3KooWHB5rTeNkQdXNJ9ynvGz8Lpnmsctt7Tvp7mrYv6bcwbPG",
		"/dns/kusama.bootnodes.polkadotters.com/tcp/30313/wss/p2p/12D3KooWHB5rTeNkQdXNJ9ynvGz8Lpnmsctt7Tvp7mrYv6bcwbPG",
		"/dns/boot-cr.gatotech.network/tcp/33200/p2p/12D3KooWRNZXf99BfzQDE1C8YhuBbuy7Sj18UEf7FNpD8egbURYD",
		"/dns/boot-cr.gatotech.network/tcp/35200/wss/p2p/12D3KooWRNZXf99BfzQDE1C8YhuBbuy7Sj18UEf7FNpD8egbURYD",
		"/dns/boot-kusama.metaspan.io/tcp/23012/p2p/12D3KooWE1tq9ZL9AAxMiUBBqy1ENmh5pwfWabnoBPMo8gFPXhn6",
		"/dns/boot-kusama.metaspan.io/tcp/23015/ws/p2p/12D3KooWE1tq9ZL9AAxMiUBBqy1ENmh5pwfWabnoBPMo8gFPXhn6",
		"/dns/boot-kusama.metaspan.io/tcp/23016/wss/p2p/12D3KooWE1tq9ZL9AAxMiUBBqy1ENmh5pwfWabnoBPMo8gFPXhn6",
		"/dns/kusama-bootnode.turboflakes.io/tcp/30305/p2p/12D3KooWR6cMhCYRhbJdqYZfzWZT6bcck3unpRLk8GBQGmHBgPwu",
		"/dns/kusama-bootnode.turboflakes.io/tcp/30405/wss/p2p/12D3KooWR6cMhCYRhbJdqYZfzWZT6bcck3unpRLk8GBQGmHBgPwu",
		"/dns/kusama-boot-ng.dwellir.com/tcp/443/wss/p2p/12D3KooWLswepVYVdCNduvWRTyNTaDMXEBcmvJdZ9Bhw3u2Jhad2",
		"/dns/kusama-boot-ng.dwellir.com/tcp/30334/p2p/12D3KooWLswepVYVdCNduvWRTyNTaDMXEBcmvJdZ9Bhw3u2Jhad2",
		"/dns/kusama-bootnode.radiumblock.com/tcp/30335/wss/p2p/12D3KooWGzKffWe7JSXeKMQeSQC5xfBafZtgBDCuBVxmwe2TJRuc",
		"/dns/kusama-bootnode.radiumblock.com/tcp/30333/p2p/12D3KooWGzKffWe7JSXeKMQeSQC5xfBafZtgBDCuBVxmwe2TJRuc",
		"/dns/ksm-bootnode.stakeworld.io/tcp/30300/p2p/12D3KooWFRin7WWVS6RgUsSpkfUHSv4tfGKnr2zJPmf1pbMv118H",
		"/dns/ksm-bootnode.stakeworld.io/tcp/30301/ws/p2p/12D3KooWFRin7WWVS6RgUsSpkfUHSv4tfGKnr2zJPmf1pbMv118H",
		"/dns/ksm-bootnode.stakeworld.io/tcp/30302/wss/p2p/12D3KooWFRin7WWVS6RgUsSpkfUHSv4tfGKnr2zJPmf1pbMv118H",
		"/dns/ksm14.rotko.net/tcp/35224/wss/p2p/12D3KooWAa5THTw8HPfnhEei23HdL8P9McBXdozG2oTtMMksjZkK",
		"/dns/ksm14.rotko.net/tcp/33224/p2p/12D3KooWAa5THTw8HPfnhEei23HdL8P9McBXdozG2oTtMMksjZkK",
		"/dns/ibp-boot-kusama.luckyfriday.io/tcp/30333/p2p/12D3KooW9vu1GWHBuxyhm7rZgD3fhGZpNajPXFexadvhujWMgwfT",
		"/dns/boot-kusama.luckyfriday.io/tcp/443/wss/p2p/12D3KooWS1Lu6DmK8YHSvkErpxpcXmk14vG6y4KVEFEkd9g62PP8",
		"/dns/ibp-boot-kusama.luckyfriday.io/tcp/30334/wss/p2p/12D3KooW9vu1GWHBuxyhm7rZgD3fhGZpNajPXFexadvhujWMgwfT",
	}

	// BootstrapPeersPolkadot extracted from:
	//   https://github.com/paritytech/polkadot-sdk/blob/master/polkadot/node/service/chain-specs/polkadot.json
	BootstrapPeersPolkadot = []string{
		"/dns/polkadot-bootnode-0.polkadot.io/tcp/30333/p2p/12D3KooWSz8r2WyCdsfWHgPyvD8GKQdJ1UAiRmrcrs8sQB3fe2KU",
		"/dns/polkadot-bootnode-0.polkadot.io/tcp/30334/ws/p2p/12D3KooWSz8r2WyCdsfWHgPyvD8GKQdJ1UAiRmrcrs8sQB3fe2KU",
		"/dns/polkadot-bootnode-0.polkadot.io/tcp/443/wss/p2p/12D3KooWSz8r2WyCdsfWHgPyvD8GKQdJ1UAiRmrcrs8sQB3fe2KU",
		"/dns/polkadot-bootnode-1.polkadot.io/tcp/30333/p2p/12D3KooWFN2mhgpkJsDBuNuE5427AcDrsib8EoqGMZmkxWwx3Md4",
		"/dns/polkadot-bootnode-1.polkadot.io/tcp/30334/ws/p2p/12D3KooWFN2mhgpkJsDBuNuE5427AcDrsib8EoqGMZmkxWwx3Md4",
		"/dns/polkadot-bootnode-1.polkadot.io/tcp/443/wss/p2p/12D3KooWFN2mhgpkJsDBuNuE5427AcDrsib8EoqGMZmkxWwx3Md4",
		"/dns/polkadot-boot.dwellir.com/tcp/30334/ws/p2p/12D3KooWKvdDyRKqUfSAaUCbYiLwKY8uK3wDWpCuy2FiDLbkPTDJ",
		"/dns/polkadot-boot.dwellir.com/tcp/443/wss/p2p/12D3KooWKvdDyRKqUfSAaUCbYiLwKY8uK3wDWpCuy2FiDLbkPTDJ",
		"/dns/boot.stake.plus/tcp/30333/p2p/12D3KooWKT4ZHNxXH4icMjdrv7EwWBkfbz5duxE5sdJKKeWFYi5n",
		"/dns/boot.stake.plus/tcp/30334/wss/p2p/12D3KooWKT4ZHNxXH4icMjdrv7EwWBkfbz5duxE5sdJKKeWFYi5n",
		"/dns/boot-node.helikon.io/tcp/7070/p2p/12D3KooWS9ZcvRxyzrSf6p63QfTCWs12nLoNKhGux865crgxVA4H",
		"/dns/boot-node.helikon.io/tcp/7072/wss/p2p/12D3KooWS9ZcvRxyzrSf6p63QfTCWs12nLoNKhGux865crgxVA4H",
		"/dns/polkadot.bootnode.amforc.com/tcp/30333/p2p/12D3KooWAsuCEVCzUVUrtib8W82Yne3jgVGhQZN3hizko5FTnDg3",
		"/dns/polkadot.bootnode.amforc.com/tcp/30334/wss/p2p/12D3KooWAsuCEVCzUVUrtib8W82Yne3jgVGhQZN3hizko5FTnDg3",
		"/dns/polkadot.bootnodes.polkadotters.com/tcp/30314/p2p/12D3KooWPAVUgBaBk6n8SztLrMk8ESByncbAfRKUdxY1nygb9zG3",
		"/dns/polkadot.bootnodes.polkadotters.com/tcp/30316/wss/p2p/12D3KooWPAVUgBaBk6n8SztLrMk8ESByncbAfRKUdxY1nygb9zG3",
		"/dns/boot-cr.gatotech.network/tcp/33100/p2p/12D3KooWK4E16jKk9nRhvC4RfrDVgcZzExg8Q3Q2G7ABUUitks1w",
		"/dns/boot-cr.gatotech.network/tcp/35100/wss/p2p/12D3KooWK4E16jKk9nRhvC4RfrDVgcZzExg8Q3Q2G7ABUUitks1w",
		"/dns/boot-polkadot.metaspan.io/tcp/13012/p2p/12D3KooWRjHFApinuqSBjoaDjQHvxwubQSpEVy5hrgC9Smvh92WF",
		"/dns/boot-polkadot.metaspan.io/tcp/13015/ws/p2p/12D3KooWRjHFApinuqSBjoaDjQHvxwubQSpEVy5hrgC9Smvh92WF",
		"/dns/boot-polkadot.metaspan.io/tcp/13016/wss/p2p/12D3KooWRjHFApinuqSBjoaDjQHvxwubQSpEVy5hrgC9Smvh92WF",
		"/dns/polkadot-bootnode.turboflakes.io/tcp/30300/p2p/12D3KooWHJBMZgt7ymAdTRtadPcGXpJw79vBGe8z53r9JMkZW7Ha",
		"/dns/polkadot-bootnode.turboflakes.io/tcp/30400/wss/p2p/12D3KooWHJBMZgt7ymAdTRtadPcGXpJw79vBGe8z53r9JMkZW7Ha",
		"/dns/polkadot-boot-ng.dwellir.com/tcp/443/wss/p2p/12D3KooWFFqjBKoSdQniRpw1Y8W6kkV7takWv1DU2ZMkaA81PYVq",
		"/dns/polkadot-boot-ng.dwellir.com/tcp/30336/p2p/12D3KooWFFqjBKoSdQniRpw1Y8W6kkV7takWv1DU2ZMkaA81PYVq",
		"/dns/polkadot-bootnode.radiumblock.com/tcp/30335/wss/p2p/12D3KooWNwWNRrPrTk4qMah1YszudMjxNw2qag7Kunhw3Ghs9ea5",
		"/dns/polkadot-bootnode.radiumblock.com/tcp/30333/p2p/12D3KooWNwWNRrPrTk4qMah1YszudMjxNw2qag7Kunhw3Ghs9ea5",
		"/dns/dot-bootnode.stakeworld.io/tcp/30310/p2p/12D3KooWAb5MyC1UJiEQJk4Hg4B2Vi3AJdqSUhTGYUqSnEqCFMFg",
		"/dns/dot-bootnode.stakeworld.io/tcp/30311/ws/p2p/12D3KooWAb5MyC1UJiEQJk4Hg4B2Vi3AJdqSUhTGYUqSnEqCFMFg",
		"/dns/dot-bootnode.stakeworld.io/tcp/30312/wss/p2p/12D3KooWAb5MyC1UJiEQJk4Hg4B2Vi3AJdqSUhTGYUqSnEqCFMFg",
		"/dns/dot14.rotko.net/tcp/35214/wss/p2p/12D3KooWPyEvPEXghnMC67Gff6PuZiSvfx3fmziKiPZcGStZ5xff",
		"/dns/dot14.rotko.net/tcp/33214/p2p/12D3KooWPyEvPEXghnMC67Gff6PuZiSvfx3fmziKiPZcGStZ5xff",
		"/dns/ibp-boot-polkadot.luckyfriday.io/tcp/30333/p2p/12D3KooWEjk6QXrZJ26fLpaajisJGHiz6WiQsR8k7mkM9GmWKnRZ",
		"/dns/ibp-boot-polkadot.luckyfriday.io/tcp/30334/wss/p2p/12D3KooWEjk6QXrZJ26fLpaajisJGHiz6WiQsR8k7mkM9GmWKnRZ",
		"/dns/boot-polkadot.luckyfriday.io/tcp/443/wss/p2p/12D3KooWAdyiVAaeGdtBt6vn5zVetwA4z4qfm9Fi2QCSykN1wTBJ",
	}

	// BootstrapPeersRococo extracted from:
	//   https://github.com/paritytech/polkadot-sdk/blob/master/polkadot/node/service/chain-specs/rococo.json
	BootstrapPeersRococo = []string{
		"/dns/rococo-bootnode-0.polkadot.io/tcp/30333/p2p/12D3KooWGikJMBmRiG5ofCqn8aijCijgfmZR5H9f53yUF3srm6Nm",
		"/dns/rococo-bootnode-1.polkadot.io/tcp/30333/p2p/12D3KooWLDfH9mHRCidrd5NfQjp7rRMUcJSEUwSvEKyu7xU2cG3d",
		"/dns/rococo-bootnode-0.polkadot.io/tcp/30334/ws/p2p/12D3KooWGikJMBmRiG5ofCqn8aijCijgfmZR5H9f53yUF3srm6Nm",
		"/dns/rococo-bootnode-1.polkadot.io/tcp/30334/ws/p2p/12D3KooWLDfH9mHRCidrd5NfQjp7rRMUcJSEUwSvEKyu7xU2cG3d",
		"/dns/rococo-bootnode-0.polkadot.io/tcp/443/wss/p2p/12D3KooWGikJMBmRiG5ofCqn8aijCijgfmZR5H9f53yUF3srm6Nm",
		"/dns/rococo-bootnode-1.polkadot.io/tcp/443/wss/p2p/12D3KooWLDfH9mHRCidrd5NfQjp7rRMUcJSEUwSvEKyu7xU2cG3d",
	}

	// BootstrapPeersWestend extracted from:
	//   https://github.com/paritytech/polkadot-sdk/blob/master/polkadot/node/service/chain-specs/westend.json
	BootstrapPeersWestend = []string{
		"/dns/westend-bootnode-0.polkadot.io/tcp/30333/p2p/12D3KooWKer94o1REDPtAhjtYR4SdLehnSrN8PEhBnZm5NBoCrMC",
		"/dns/westend-bootnode-0.polkadot.io/tcp/30334/ws/p2p/12D3KooWKer94o1REDPtAhjtYR4SdLehnSrN8PEhBnZm5NBoCrMC",
		"/dns/westend-bootnode-0.polkadot.io/tcp/443/wss/p2p/12D3KooWKer94o1REDPtAhjtYR4SdLehnSrN8PEhBnZm5NBoCrMC",
		"/dns/westend-bootnode-1.polkadot.io/tcp/30333/p2p/12D3KooWPVPzs42GvRBShdUMtFsk4SvnByrSdWqb6aeAAHvLMSLS",
		"/dns/westend-bootnode-1.polkadot.io/tcp/30334/ws/p2p/12D3KooWPVPzs42GvRBShdUMtFsk4SvnByrSdWqb6aeAAHvLMSLS",
		"/dns/westend-bootnode-1.polkadot.io/tcp/443/wss/p2p/12D3KooWPVPzs42GvRBShdUMtFsk4SvnByrSdWqb6aeAAHvLMSLS",
		"/dns/boot.stake.plus/tcp/32333/p2p/12D3KooWK8fjVoSvMq5copQYMsdYreSGPGgcMbGMgbMDPfpf3sm7",
		"/dns/boot.stake.plus/tcp/32334/wss/p2p/12D3KooWK8fjVoSvMq5copQYMsdYreSGPGgcMbGMgbMDPfpf3sm7",
		"/dns/boot-node.helikon.io/tcp/7080/p2p/12D3KooWRFDPyT8vA8mLzh6dJoyujn4QNjeqi6Ch79eSMz9beKXC",
		"/dns/boot-node.helikon.io/tcp/7082/wss/p2p/12D3KooWRFDPyT8vA8mLzh6dJoyujn4QNjeqi6Ch79eSMz9beKXC",
		"/dns/westend.bootnode.amforc.com/tcp/30333/p2p/12D3KooWJ5y9ZgVepBQNW4aabrxgmnrApdVnscqgKWiUu4BNJbC8",
		"/dns/westend.bootnode.amforc.com/tcp/30334/wss/p2p/12D3KooWJ5y9ZgVepBQNW4aabrxgmnrApdVnscqgKWiUu4BNJbC8",
		"/dns/westend.bootnodes.polkadotters.com/tcp/30308/p2p/12D3KooWHPHb64jXMtSRJDrYFATWeLnvChL8NtWVttY67DCH1eC5",
		"/dns/westend.bootnodes.polkadotters.com/tcp/30310/wss/p2p/12D3KooWHPHb64jXMtSRJDrYFATWeLnvChL8NtWVttY67DCH1eC5",
		"/dns/boot-cr.gatotech.network/tcp/33300/p2p/12D3KooWQGR1vUhoy6mvQorFp3bZFn6NNezhQZ6NWnVV7tpFgoPd",
		"/dns/boot-cr.gatotech.network/tcp/35300/wss/p2p/12D3KooWQGR1vUhoy6mvQorFp3bZFn6NNezhQZ6NWnVV7tpFgoPd",
		"/dns/boot-westend.metaspan.io/tcp/33012/p2p/12D3KooWNTau7iG4G9cUJSwwt2QJP1W88pUf2SgqsHjRU2RL8pfa",
		"/dns/boot-westend.metaspan.io/tcp/33015/ws/p2p/12D3KooWNTau7iG4G9cUJSwwt2QJP1W88pUf2SgqsHjRU2RL8pfa",
		"/dns/boot-westend.metaspan.io/tcp/33016/wss/p2p/12D3KooWNTau7iG4G9cUJSwwt2QJP1W88pUf2SgqsHjRU2RL8pfa",
		"/dns/westend-bootnode.turboflakes.io/tcp/30310/p2p/12D3KooWJvPDCZmReU46ghpCMJCPVUvUCav4WQdKtXQhZgJdH6tZ",
		"/dns/westend-bootnode.turboflakes.io/tcp/30410/wss/p2p/12D3KooWJvPDCZmReU46ghpCMJCPVUvUCav4WQdKtXQhZgJdH6tZ",
		"/dns/westend-boot-ng.dwellir.com/tcp/443/wss/p2p/12D3KooWJifoDhCL3swAKt7MWhFb7wLRFD9oG33AL3nAathmU24x",
		"/dns/westend-boot-ng.dwellir.com/tcp/30335/p2p/12D3KooWJifoDhCL3swAKt7MWhFb7wLRFD9oG33AL3nAathmU24x",
		"/dns/westend-bootnode.radiumblock.com/tcp/30335/wss/p2p/12D3KooWJBowJuX1TaWNWHt8Dz8z44BoCZunLCfFqxA2rLTn6TBD",
		"/dns/westend-bootnode.radiumblock.com/tcp/30333/p2p/12D3KooWJBowJuX1TaWNWHt8Dz8z44BoCZunLCfFqxA2rLTn6TBD",
		"/dns/wnd-bootnode.stakeworld.io/tcp/30320/p2p/12D3KooWBYdKipcNbrV5rCbgT5hco8HMLME7cE9hHC3ckqCKDuzP",
		"/dns/wnd-bootnode.stakeworld.io/tcp/30321/ws/p2p/12D3KooWBYdKipcNbrV5rCbgT5hco8HMLME7cE9hHC3ckqCKDuzP",
		"/dns/wnd-bootnode.stakeworld.io/tcp/30322/wss/p2p/12D3KooWBYdKipcNbrV5rCbgT5hco8HMLME7cE9hHC3ckqCKDuzP",
		"/dns/wnd14.rotko.net/tcp/35234/wss/p2p/12D3KooWLK8Zj1uZ46phU3vQwiDVda8tB76S8J26rXZQLHpwWkDJ",
		"/dns/wnd14.rotko.net/tcp/33234/p2p/12D3KooWLK8Zj1uZ46phU3vQwiDVda8tB76S8J26rXZQLHpwWkDJ",
		"/dns/ibp-boot-westend.luckyfriday.io/tcp/30333/p2p/12D3KooWDg1YEytdwFFNWroFj6gio4YFsMB3miSbHKgdpJteUMB9",
		"/dns/ibp-boot-westend.luckyfriday.io/tcp/30334/wss/p2p/12D3KooWDg1YEytdwFFNWroFj6gio4YFsMB3miSbHKgdpJteUMB9",
	}

	// BootstrapPeersCelestia extracted from:
	//   https://github.com/celestiaorg/celestia-node/blob/3fde8c9acb90bb9d4d1abbdc2c5917d7f3693388/nodebuilder/p2p/bootstrap.go#L39
	BootstrapPeersCelestia = []string{
		"/dns4/da-bridge-1.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWSqZaLcn5Guypo2mrHr297YPJnV8KMEMXNjs3qAS8msw8",
		"/dns4/da-bridge-2.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWQpuTFELgsUypqp9N4a1rKBccmrmQVY8Em9yhqppTJcXf",
		"/dns4/da-bridge-3.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWSGa4huD6ts816navn7KFYiStBiy5LrBQH1HuEahk4TzQ",
		"/dns4/da-bridge-4.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWHBXCmXaUNat6ooynXG837JXPsZpSTeSzZx6DpgNatMmR",
		"/dns4/da-bridge-5.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWDGTBK1a2Ru1qmnnRwP6Dmc44Zpsxi3xbgFk7ATEPfmEU",
		"/dns4/da-bridge-6.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWLTUFyf3QEGqYkHWQS2yCtuUcL78vnKBdXU5gABM1YDeH",
		"/dns4/da-full-1.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWKZCMcwGCYbL18iuw3YVpAZoyb1VBGbx9Kapsjw3soZgr",
		"/dns4/da-full-2.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWE3fmRtHgfk9DCuQFfY3H3JYEnTU3xZozv1Xmo8KWrWbK",
		"/dns4/da-full-3.celestia-bootstrap.net/tcp/2121/p2p/12D3KooWK6Ftsd4XsWCsQZgZPNhTrE5urwmkoo5P61tGvnKmNVyv",
	}

	// BootstrapPeersArabica extracted from:
	//   https://github.com/celestiaorg/celestia-node/blob/3fde8c9acb90bb9d4d1abbdc2c5917d7f3693388/nodebuilder/p2p/bootstrap.go#L51
	BootstrapPeersArabica = []string{
		"/dns4/da-bridge.celestia-arabica-10.com/tcp/2121/p2p/12D3KooWM3e9MWtyc8GkP8QRt74Riu17QuhGfZMytB2vq5NwkWAu",
		"/dns4/da-bridge-2.celestia-arabica-10.com/tcp/2121/p2p/12D3KooWKj8mcdiBGxQRe1jqhaMnh2tGoC3rPDmr5UH2q8H4WA9M",
		"/dns4/da-full-1.celestia-arabica-10.com/tcp/2121/p2p/12D3KooWBWkgmN7kmJSFovVrCjkeG47FkLGq7yEwJ2kEqNKCsBYk",
		"/dns4/da-full-2.celestia-arabica-10.com/tcp/2121/p2p/12D3KooWRByRF67a2kVM2j4MP5Po3jgTw7H2iL2Spu8aUwPkrRfP",
	}

	// BootstrapPeersMocha extracted from:
	//   https://github.com/celestiaorg/celestia-node/blob/3fde8c9acb90bb9d4d1abbdc2c5917d7f3693388/nodebuilder/p2p/bootstrap.go#L57
	BootstrapPeersMocha = []string{
		"/dns4/da-bridge-mocha-4.celestia-mocha.com/tcp/2121/p2p/12D3KooWCBAbQbJSpCpCGKzqz3rAN4ixYbc63K68zJg9aisuAajg",
		"/dns4/da-bridge-mocha-4-2.celestia-mocha.com/tcp/2121/p2p/12D3KooWK6wJkScGQniymdWtBwBuU36n6BRXp9rCDDUD6P5gJr3G",
		"/dns4/da-full-1-mocha-4.celestia-mocha.com/tcp/2121/p2p/12D3KooWCUHPLqQXZzpTx1x3TAsdn3vYmTNDhzg66yG8hqoxGGN8",
		"/dns4/da-full-2-mocha-4.celestia-mocha.com/tcp/2121/p2p/12D3KooWR6SHsXPkkvhCRn6vp1RqSefgaT1X1nMNvrVjU2o3GoYy",
	}

	// BootstrapPeersBlockspaceRace extracted from:
	//   https://github.com/celestiaorg/celestia-node/blob/9c0a5fb0626ada6e6cdb8bcd816d01a3aa5043ad/nodebuilder/p2p/bootstrap.go#L39
	BootstrapPeersBlockspaceRace = []string{
		"/dns4/bootstr-incent-3.celestia.tools/tcp/2121/p2p/12D3KooWNzdKcHagtvvr6qtjcPTAdCN6ZBiBLH8FBHbihxqu4GZx",
		"/dns4/bootstr-incent-2.celestia.tools/tcp/2121/p2p/12D3KooWNJZyWeCsrKxKrxsNM1RVL2Edp77svvt7Cosa63TggC9m",
		"/dns4/bootstr-incent-1.celestia.tools/tcp/2121/p2p/12D3KooWBtxdBzToQwnS4ySGpph9PtGmmjEyATkgX3PfhAo4xmf7",
	}

	// BootstrapPeersEthereumConsensus extracted from:
	//   https://github.com/eth-clients/eth2-networks/blob/master/shared/mainnet/bootstrap_nodes.txt
	BootstrapPeersEthereumConsensus = []string{
		// Teku team's bootnodes
		"enr:-KG4QMOEswP62yzDjSwWS4YEjtTZ5PO6r65CPqYBkgTTkrpaedQ8uEUo1uMALtJIvb2w_WWEVmg5yt1UAuK1ftxUU7QDhGV0aDKQu6TalgMAAAD__________4JpZIJ2NIJpcIQEnfA2iXNlY3AyNTZrMaEDfol8oLr6XJ7FsdAYE7lpJhKMls4G_v6qQOGKJUWGb_uDdGNwgiMog3VkcIIjKA",
		"enr:-KG4QF4B5WrlFcRhUU6dZETwY5ZzAXnA0vGC__L1Kdw602nDZwXSTs5RFXFIFUnbQJmhNGVU6OIX7KVrCSTODsz1tK4DhGV0aDKQu6TalgMAAAD__________4JpZIJ2NIJpcIQExNYEiXNlY3AyNTZrMaECQmM9vp7KhaXhI-nqL_R0ovULLCFSFTa9CPPSdb1zPX6DdGNwgiMog3VkcIIjKA",
		// Prylab team's bootnodes
		"enr:-Ku4QImhMc1z8yCiNJ1TyUxdcfNucje3BGwEHzodEZUan8PherEo4sF7pPHPSIB1NNuSg5fZy7qFsjmUKs2ea1Whi0EBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhBLf22SJc2VjcDI1NmsxoQOVphkDqal4QzPMksc5wnpuC3gvSC8AfbFOnZY_On34wIN1ZHCCIyg",
		"enr:-Ku4QP2xDnEtUXIjzJ_DhlCRN9SN99RYQPJL92TMlSv7U5C1YnYLjwOQHgZIUXw6c-BvRg2Yc2QsZxxoS_pPRVe0yK8Bh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhBLf22SJc2VjcDI1NmsxoQMeFF5GrS7UZpAH2Ly84aLK-TyvH-dRo0JM1i8yygH50YN1ZHCCJxA",
		"enr:-Ku4QPp9z1W4tAO8Ber_NQierYaOStqhDqQdOPY3bB3jDgkjcbk6YrEnVYIiCBbTxuar3CzS528d2iE7TdJsrL-dEKoBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhBLf22SJc2VjcDI1NmsxoQMw5fqqkw2hHC4F5HZZDPsNmPdB1Gi8JPQK7pRc9XHh-oN1ZHCCKvg",
		// Lighthouse team's bootnodes
		"enr:-Le4QPUXJS2BTORXxyx2Ia-9ae4YqA_JWX3ssj4E_J-3z1A-HmFGrU8BpvpqhNabayXeOZ2Nq_sbeDgtzMJpLLnXFgAChGV0aDKQtTA_KgEAAAAAIgEAAAAAAIJpZIJ2NIJpcISsaa0Zg2lwNpAkAIkHAAAAAPA8kv_-awoTiXNlY3AyNTZrMaEDHAD2JKYevx89W0CcFJFiskdcEzkH_Wdv9iW42qLK79ODdWRwgiMohHVkcDaCI4I",
		"enr:-Le4QLHZDSvkLfqgEo8IWGG96h6mxwe_PsggC20CL3neLBjfXLGAQFOPSltZ7oP6ol54OvaNqO02Rnvb8YmDR274uq8ChGV0aDKQtTA_KgEAAAAAIgEAAAAAAIJpZIJ2NIJpcISLosQxg2lwNpAqAX4AAAAAAPA8kv_-ax65iXNlY3AyNTZrMaEDBJj7_dLFACaxBfaI8KZTh_SSJUjhyAyfshimvSqo22WDdWRwgiMohHVkcDaCI4I",
		"enr:-Le4QH6LQrusDbAHPjU_HcKOuMeXfdEB5NJyXgHWFadfHgiySqeDyusQMvfphdYWOzuSZO9Uq2AMRJR5O4ip7OvVma8BhGV0aDKQtTA_KgEAAAAAIgEAAAAAAIJpZIJ2NIJpcISLY9ncg2lwNpAkAh8AgQIBAAAAAAAAAAmXiXNlY3AyNTZrMaECDYCZTZEksF-kmgPholqgVt8IXr-8L7Nu7YrZ7HUpgxmDdWRwgiMohHVkcDaCI4I",
		"enr:-Le4QIqLuWybHNONr933Lk0dcMmAB5WgvGKRyDihy1wHDIVlNuuztX62W51voT4I8qD34GcTEOTmag1bcdZ_8aaT4NUBhGV0aDKQtTA_KgEAAAAAIgEAAAAAAIJpZIJ2NIJpcISLY04ng2lwNpAkAh8AgAIBAAAAAAAAAA-fiXNlY3AyNTZrMaEDscnRV6n1m-D9ID5UsURk0jsoKNXt1TIrj8uKOGW6iluDdWRwgiMohHVkcDaCI4I",
		// EF bootnodes
		"enr:-Ku4QHqVeJ8PPICcWk1vSn_XcSkjOkNiTg6Fmii5j6vUQgvzMc9L1goFnLKgXqBJspJjIsB91LTOleFmyWWrFVATGngBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhAMRHkWJc2VjcDI1NmsxoQKLVXFOhp2uX6jeT0DvvDpPcU8FWMjQdR4wMuORMhpX24N1ZHCCIyg",
		"enr:-Ku4QG-2_Md3sZIAUebGYT6g0SMskIml77l6yR-M_JXc-UdNHCmHQeOiMLbylPejyJsdAPsTHJyjJB2sYGDLe0dn8uYBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhBLY-NyJc2VjcDI1NmsxoQORcM6e19T1T9gi7jxEZjk_sjVLGFscUNqAY9obgZaxbIN1ZHCCIyg",
		"enr:-Ku4QPn5eVhcoF1opaFEvg1b6JNFD2rqVkHQ8HApOKK61OIcIXD127bKWgAtbwI7pnxx6cDyk_nI88TrZKQaGMZj0q0Bh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhDayLMaJc2VjcDI1NmsxoQK2sBOLGcUb4AwuYzFuAVCaNHA-dy24UuEKkeFNgCVCsIN1ZHCCIyg",
		"enr:-Ku4QEWzdnVtXc2Q0ZVigfCGggOVB2Vc1ZCPEc6j21NIFLODSJbvNaef1g4PxhPwl_3kax86YPheFUSLXPRs98vvYsoBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhDZBrP2Jc2VjcDI1NmsxoQM6jr8Rb1ktLEsVcKAPa08wCsKUmvoQ8khiOl_SLozf9IN1ZHCCIyg",
		// Nimbus team's bootnodes
		"enr:-LK4QA8FfhaAjlb_BXsXxSfiysR7R52Nhi9JBt4F8SPssu8hdE1BXQQEtVDC3qStCW60LSO7hEsVHv5zm8_6Vnjhcn0Bh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhAN4aBKJc2VjcDI1NmsxoQJerDhsJ-KxZ8sHySMOCmTO6sHM3iCFQ6VMvLTe948MyYN0Y3CCI4yDdWRwgiOM",
		"enr:-LK4QKWrXTpV9T78hNG6s8AM6IO4XH9kFT91uZtFg1GcsJ6dKovDOr1jtAAFPnS2lvNltkOGA9k29BUN7lFh_sjuc9QBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhANAdd-Jc2VjcDI1NmsxoQLQa6ai7y9PMN5hpLe5HmiJSlYzMuzP7ZhwRiwHvqNXdoN0Y3CCI4yDdWRwgiOM",
	}

	// BootstrapPeersEthereumExecution extracted from:
	//   https://github.com/ethereum/go-ethereum/blob/f04e5bde7487ce554930187e766164b18c37d867/params/bootnodes.go#L23
	BootstrapPeersEthereumExecution = []string{
		"enode://d860a01f9722d78051619d1e2351aba3f43f943f6f00718d1b9baa4101932a1f5011f16bb2b1bb35db20d6fe28fa0bf09636d26a87d31de9ec6203eeedb1f666@18.138.108.67:30303", // bootnode-aws-ap-southeast-1-001
		"enode://22a8232c3abc76a16ae9d6c3b164f98775fe226f0917b0ca871128a74a8e9630b458460865bab457221f1d448dd9791d24c4e5d88786180ac185df813a68d4de@3.209.45.79:30303",   // bootnode-aws-us-east-1-001
		"enode://2b252ab6a1d0f971d9722cb839a42cb81db019ba44c08754628ab4a823487071b5695317c8ccd085219c3a03af063495b2f1da8d18218da2d6a82981b45e6ffc@65.108.70.101:30303", // bootnode-hetzner-hel
		"enode://4aeb4ab6c14b23e2c4cfdce879c04b0748a20d8e9b59e25ded2a08143e265c6c25936e74cbc8e641e3312ca288673d91f2f93f8e277de3cfa444ecdaaf982052@157.90.35.166:30303", // bootnode-hetzner-fsn
	}

	// BootstrapPeersHolesky extracted from:
	//   https://github.com/eth-clients/holesky/blob/main/custom_config_data/bootstrap_nodes.txt
	BootstrapPeersHolesky = []string{
		"enr:-Ku4QFo-9q73SspYI8cac_4kTX7yF800VXqJW4Lj3HkIkb5CMqFLxciNHePmMt4XdJzHvhrCC5ADI4D_GkAsxGJRLnQBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpAhnTT-AQFwAP__________gmlkgnY0gmlwhLKAiOmJc2VjcDI1NmsxoQORcM6e19T1T9gi7jxEZjk_sjVLGFscUNqAY9obgZaxbIN1ZHCCIyk",
		"enr:-Ku4QPG7F72mbKx3gEQEx07wpYYusGDh-ni6SNkLvOS-hhN-BxIggN7tKlmalb0L5JPoAfqD-akTZ-gX06hFeBEz4WoBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpAhnTT-AQFwAP__________gmlkgnY0gmlwhJK-DYCJc2VjcDI1NmsxoQKLVXFOhp2uX6jeT0DvvDpPcU8FWMjQdR4wMuORMhpX24N1ZHCCIyk",
		"enr:-LK4QPxe-mDiSOtEB_Y82ozvxn9aQM07Ui8A-vQHNgYGMMthfsfOabaaTHhhJHFCBQQVRjBww_A5bM1rf8MlkJU_l68Eh2F0dG5ldHOIAADAAAAAAACEZXRoMpBpt9l0BAFwAAABAAAAAAAAgmlkgnY0gmlwhLKAiOmJc2VjcDI1NmsxoQJu6T9pclPObAzEVQ53DpVQqjadmVxdTLL-J3h9NFoCeIN0Y3CCIyiDdWRwgiMo",
		"enr:-Ly4QGbOw4xNel5EhmDsJJ-QhC9XycWtsetnWoZ0uRy381GHdHsNHJiCwDTOkb3S1Ade0SFQkWJX_pgb3g8Jfh93rvMBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpBpt9l0BAFwAAABAAAAAAAAgmlkgnY0gmlwhJK-DYCJc2VjcDI1NmsxoQOxKv9sv3zKF8GDewgFGGHKP5HCZZpPpTrwl9eXKAWGxIhzeW5jbmV0cwCDdGNwgiMog3VkcIIjKA",
		"enr:-LS4QG0uV4qvcpJ-HFDJRGBmnlD3TJo7yc4jwK8iP7iKaTlfQ5kZvIDspLMJhk7j9KapuL9yyHaZmwTEZqr10k9XumyCEcmHYXR0bmV0c4gAAAAABgAAAIRldGgykGm32XQEAXAAAAEAAAAAAACCaWSCdjSCaXCErK4j-YlzZWNwMjU2azGhAgfWRBEJlb7gAhXIB5ePmjj2b8io0UpEenq1Kl9cxStJg3RjcIIjKIN1ZHCCIyg",
		"enr:-Le4QLoE1wFHSlGcm48a9ZESb_MRLqPPu6G0vHqu4MaUcQNDHS69tsy-zkN0K6pglyzX8m24mkb-LtBcbjAYdP1uxm4BhGV0aDKQabfZdAQBcAAAAQAAAAAAAIJpZIJ2NIJpcIQ5gR6Wg2lwNpAgAUHQBwEQAAAAAAAAADR-iXNlY3AyNTZrMaEDPMSNdcL92uNIyCsS177Z6KTXlbZakQqxv3aQcWawNXeDdWRwgiMohHVkcDaCI4I",
	}

	// BootstrapPeersGoldbergLightClient Avail network extracted from:
	//   https://github.com/availproject/light-client-lib/blob/e5cb20f092951e15436bc21a2ac1f97c10d6a800/src/types.rs#L586C6-L586C72
	BootstrapPeersGoldbergLightClient = []string{
		"/dns/bootnode.1.lightclient.goldberg.avail.tools/tcp/37000/p2p/12D3KooWBkLsNGaD3SpMaRWtAmWVuiZg1afdNSPbtJ8M8r9ArGRT",
	}

	// BootstrapPeersGoldbergFullNode Avail network extracted from:
	//   https://github.com/availproject/avail/blob/e976c560ee3f8c7d36d9fbec8e98c29283f1aa34/misc/genesis/testnet.goldberg.chain.json#L5
	BootstrapPeersGoldbergFullNode = []string{
		"/dns/bootnode-001.goldberg.avail.tools/tcp/30333/p2p/12D3KooWCVqFvrP3UJ1S338Gb8SHvEQ1xpENLb45Dbynk4hu1XGN",
		"/dns/bootnode-002.goldberg.avail.tools/tcp/30333/p2p/12D3KooWD6sWeWCG5Z1qhejhkPk9Rob5h75wYmPB6MUoPo7br58m",
		"/dns/bootnode-003.goldberg.avail.tools/tcp/30333/p2p/12D3KooWMR9ZoAVWJv6ahraVzUCfacNbFKk7ABoWxVL3fJ3XXGDw",
		"/dns/bootnode-004.goldberg.avail.tools/tcp/30333/p2p/12D3KooWMuyLE3aPQ82HTWuPUCjiP764ebQrZvGUzxrYGuXWZJZV",
		"/dns/bootnode-005.goldberg.avail.tools/tcp/30333/p2p/12D3KooWKJwbdcZ7QWcPLHy3EJ1UiffaLGnNBMffeK8AqRVWBZA1",
		"/dns/bootnode-006.goldberg.avail.tools/tcp/30333/p2p/12D3KooWM8AaHDH8SJvg6bq4CGQyHvW2LH7DCHbdv633dsrti7i5",
	}
)

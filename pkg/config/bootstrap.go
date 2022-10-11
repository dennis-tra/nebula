package config

import ma "github.com/multiformats/go-multiaddr"

var (
	BootstrapPeersFilecoinMaddrs []ma.Multiaddr
	BootstrapPeersFilecoin       = []string{
		"/dns4/bootstrap-0.mainnet.filops.net/tcp/1347/p2p/12D3KooWCVe8MmsEMes2FzgTpt9fXtmCY7wrq91GRiaC8PHSCCBj",
		"/dns4/bootstrap-1.mainnet.filops.net/tcp/1347/p2p/12D3KooWCwevHg1yLCvktf2nvLu7L9894mcrJR4MsBCcm4syShVc",
		"/dns4/bootstrap-2.mainnet.filops.net/tcp/1347/p2p/12D3KooWEWVwHGn2yR36gKLozmb4YjDJGerotAPGxmdWZx2nxMC4",
		"/dns4/bootstrap-3.mainnet.filops.net/tcp/1347/p2p/12D3KooWKhgq8c7NQ9iGjbyK7v7phXvG6492HQfiDaGHLHLQjk7R",
		"/dns4/bootstrap-4.mainnet.filops.net/tcp/1347/p2p/12D3KooWL6PsFNPhYftrJzGgF5U18hFoaVhfGk7xwzD8yVrHJ3Uc",
		"/dns4/bootstrap-5.mainnet.filops.net/tcp/1347/p2p/12D3KooWLFynvDQiUpXoHroV1YxKHhPJgysQGH2k3ZGwtWzR4dFH",
		"/dns4/bootstrap-6.mainnet.filops.net/tcp/1347/p2p/12D3KooWP5MwCiqdMETF9ub1P3MbCvQCcfconnYHbWg6sUJcDRQQ",
		"/dns4/bootstrap-7.mainnet.filops.net/tcp/1347/p2p/12D3KooWRs3aY1p3juFjPy8gPN95PEQChm2QKGUCAdcDCC4EBMKf",
		"/dns4/bootstrap-8.mainnet.filops.net/tcp/1347/p2p/12D3KooWScFR7385LTyR4zU1bYdzSiiAb5rnNABfVahPvVSzyTkR",
		"/dns4/lotus-bootstrap.ipfsforce.com/tcp/41778/p2p/12D3KooWGhufNmZHF3sv48aQeS13ng5XVJZ9E6qy2Ms4VzqeUsHk",
		"/dns4/bootstrap-0.starpool.in/tcp/12757/p2p/12D3KooWGHpBMeZbestVEWkfdnC9u7p6uFHXL1n7m1ZBqsEmiUzz",
		"/dns4/bootstrap-1.starpool.in/tcp/12757/p2p/12D3KooWQZrGH1PxSNZPum99M1zNvjNFM33d1AAu5DcvdHptuU7u",
		"/dns4/node.glif.io/tcp/1235/p2p/12D3KooWBF8cpp65hp2u9LK5mh19x67ftAam84z9LsfaquTDSBpt",
		"/dns4/bootstrap-0.ipfsmain.cn/tcp/34721/p2p/12D3KooWQnwEGNqcM2nAcPtRR9rAX8Hrg4k9kJLCHoTR5chJfz6d",
		"/dns4/bootstrap-1.ipfsmain.cn/tcp/34723/p2p/12D3KooWMKxMkD5DMpSWsW7dBddKxKT7L2GgbNuckz9otxvkvByP",
	}

	BootstrapPeersKusamaMaddrs []ma.Multiaddr
	BootstrapPeersKusama       = []string{
		"/dns/p2p.cc3-0.kusama.network/tcp/30100/p2p/12D3KooWDgtynm4S9M3m6ZZhXYu2RrWKdvkCSScc25xKDVSg1Sjd",
		"/dns/p2p.cc3-1.kusama.network/tcp/30100/p2p/12D3KooWNpGriWPmf621Lza9UWU9eLLBdCFaErf6d4HSK7Bcqnv4",
		"/dns/p2p.cc3-2.kusama.network/tcp/30100/p2p/12D3KooWLmLiB4AenmN2g2mHbhNXbUcNiGi99sAkSk1kAQedp8uE",
		"/dns/p2p.cc3-3.kusama.network/tcp/30100/p2p/12D3KooWEGHw84b4hfvXEfyq4XWEmWCbRGuHMHQMpby4BAtZ4xJf",
		"/dns/p2p.cc3-4.kusama.network/tcp/30100/p2p/12D3KooWF9KDPRMN8WpeyXhEeURZGP8Dmo7go1tDqi7hTYpxV9uW",
		"/dns/p2p.cc3-5.kusama.network/tcp/30100/p2p/12D3KooWDiwMeqzvgWNreS9sV1HW3pZv1PA7QGA7HUCo7FzN5gcA",
		"/dns/kusama-bootnode-0.paritytech.net/tcp/30333/p2p/12D3KooWSueCPH3puP2PcvqPJdNaDNF3jMZjtJtDiSy35pWrbt5h",
		"/dns/kusama-bootnode-0.paritytech.net/tcp/30334/ws/p2p/12D3KooWSueCPH3puP2PcvqPJdNaDNF3jMZjtJtDiSy35pWrbt5h",
		"/dns/kusama-bootnode-1.paritytech.net/tcp/30333/p2p/12D3KooWQKqane1SqWJNWMQkbia9qiMWXkcHtAdfW5eVF8hbwEDw",
	}

	BootstrapPeersPolkadotMaddrs []ma.Multiaddr
	BootstrapPeersPolkadot       = []string{
		"/dns/p2p.cc1-0.polkadot.network/tcp/30100/p2p/12D3KooWEdsXX9657ppNqqrRuaCHFvuNemasgU5msLDwSJ6WqsKc",
		"/dns/p2p.cc1-1.polkadot.network/tcp/30100/p2p/12D3KooWAtx477KzC8LwqLjWWUG6WF4Gqp2eNXmeqAG98ehAMWYH",
		"/dns/p2p.cc1-2.polkadot.network/tcp/30100/p2p/12D3KooWAGCCPZbr9UWGXPtBosTZo91Hb5M3hU8v6xbKgnC5LVao",
		"/dns/p2p.cc1-3.polkadot.network/tcp/30100/p2p/12D3KooWJ4eyPowiVcPU46pXuE2cDsiAmuBKXnFcFPapm4xKFdMJ",
		"/dns/p2p.cc1-4.polkadot.network/tcp/30100/p2p/12D3KooWNMUcqwSj38oEq1zHeGnWKmMvrCFnpMftw7JzjAtRj2rU",
		"/dns/p2p.cc1-5.polkadot.network/tcp/30100/p2p/12D3KooWDs6LnpmWDWgZyGtcLVr3E75CoBxzg1YZUPL5Bb1zz6fM",
		"/dns/cc1-0.parity.tech/tcp/30333/p2p/12D3KooWSz8r2WyCdsfWHgPyvD8GKQdJ1UAiRmrcrs8sQB3fe2KU",
		"/dns/cc1-1.parity.tech/tcp/30333/p2p/12D3KooWFN2mhgpkJsDBuNuE5427AcDrsib8EoqGMZmkxWwx3Md4",
	}
)

func init() {
	for _, maddrStr := range BootstrapPeersFilecoin {
		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			panic(err)
		}
		BootstrapPeersFilecoinMaddrs = append(BootstrapPeersFilecoinMaddrs, maddr)
	}

	for _, maddrStr := range BootstrapPeersKusama {
		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			panic(err)
		}
		BootstrapPeersKusamaMaddrs = append(BootstrapPeersKusamaMaddrs, maddr)
	}

	for _, maddrStr := range BootstrapPeersPolkadot {
		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			panic(err)
		}
		BootstrapPeersPolkadotMaddrs = append(BootstrapPeersPolkadotMaddrs, maddr)
	}
}

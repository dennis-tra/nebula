# GeoIP

This part of the repository deals with analyzing the IP addresses that are part of the IPFS/Filecoin networks. For that we're using the GeoLite2 database by [maxmind.com](https://www.maxmind.com/en/home).

The current analysis is based on the following database:

- Name: `GeoLite2-Country-CSV_20210706.zip`
- SHA256: `60e924400d3b20aff877598669395a47e0ba81063fc7c1b46893a5544dd34840`

The un-gzipped database weighs 4.2MB while the gzipped one only 2.2MB. So I'm only checking in the gzipped one. To run the analysis scripts you need to unzip the `GeoLite2-Country_20210706.tar.gz` file.

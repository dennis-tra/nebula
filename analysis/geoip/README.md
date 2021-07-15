# GeoIP
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)

This part of the repository contains analysis scripts in the form of Jupyter notebooks that map IP-Adresses to their respective countries.

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Prerequisites](#prerequisites)
- [Development](#development)
- [Database](#database)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [Support](#support)

## Prerequisites

- [`poetry`](https://python-poetry.org/) - tested with `v1.1.7`

## Development

When you cloned this repository run:

```shell
poetry install
```

to install all required dependencies. Then jump in the poetry shell that has all these dependencies available and start the Jupyter Notebook by running:

```shell
$ poetry shell
Spawning shell within /Users/username/Library/Caches/pypoetry/virtualenvs/geoip-a33IKhDm-py3.9
```

```shell
$ jupyter notebook
[W 19:01:27.226 NotebookApp] WARNING: The notebook server is listening on all IP addresses and not using encryption. This is not recommended.
[I 19:01:27.244 NotebookApp] Serving notebooks from local directory: /Users/path/to/clone/reopo
[I 19:01:27.244 NotebookApp] Jupyter Notebook 6.4.0 is running at:
[I 19:01:27.244 NotebookApp] http://hostname.local:8888/
[I 19:01:27.244 NotebookApp] Use Control-C to stop this server and shut down all kernels (twice to skip confirmation).
```

## Database

The `GeoLite2` subdirectory contains the [Maxmind](https://www.maxmind.com/en/home) database that powers the IP adress to country mapping. The database can be downloaded from [here](https://www.maxmind.com/en/accounts/579392/geoip/downloads):

- **Name:** `GeoLite2-Country-CSV_20210706.zip`
- **SHA256:** `60e924400d3b20aff877598669395a47e0ba81063fc7c1b46893a5544dd34840`

## Maintainers

[@dennis-tra](https://github.com/dennis-tra).

## Contributing

Feel free to dive in! [Open an issue](https://github.com/dennis-tra/nebula/issues/new) or submit PRs.

## Support

It would really make my day if you supported this project through [Buy Me A Coffee](https://www.buymeacoffee.com/dennistra).

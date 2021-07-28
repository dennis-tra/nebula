# Nebula-Crawler Deployment

## Docker Compose

From the repo root directory run:
```shell
docker build . -t dennis-tra/nebula-crawler:latest
```
to build a docker image of nebula.

Then run:
```shell
docker compose up
```

This will start:

1. `nebula`
   - monitoring mode
   - IPFS bootstrap nodes
2. `postgres`
   - exposed port at `5432`
3. `prometheus`
   - exposed port at `9090`
4. `grafana` 
   - exposed port at `3000`
   - provisioned postgres and prometheus datasources
   - provisioned dashboard for IPFS

You can log in to Grafana at `localhost:3000` with:
```shell
USER: admin
PASWORD: admin
```

To start a crawl run `./crawl.sh` or:

```shell
docker run \
  --network nebula \
  --name nebula_crawler \
  --hostname nebula_crawler \
  dennis-tra/nebula-crawler:latest \
  nebula --db-host=postgres crawl
```

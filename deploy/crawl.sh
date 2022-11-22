docker run \
  --network nebula \
  --name nebula_crawler \
  --hostname nebula_crawler \
  dennis-tra/nebula-crawler:sha-ce5c756 \
  nebula --db-host=postgres crawl

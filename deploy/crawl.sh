docker run \
  --network nebula \
  --name nebula_crawler \
  --hostname nebula_crawler \
  dennis-tra/nebula-crawler:latest \
  nebula --db-host=postgres crawl

# Churn
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)

This part of the repository contains analysis scripts that evaluate the churn rate based on the observed session lengths of the crawler.

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Prerequisites](#prerequisites)
- [Development](#development)
- [Database](#database)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [Support](#support)

## Prerequisites

TODO

## Development

TODO

## Database

To generate the session data csv file use the query:

```sql
COPY ( select peer_id, multi_addresses, EXTRACT(EPOCH FROM min_duration) as min_duration_s, EXTRACT(EPOCH FROM max_duration) as max_duration_s from peers as p inner join sessions as s on p.id = s.peer_id) to '/output/path/sessions.csv' DELIMITER ',' CSV HEADER;
```

## Maintainers

[@dennis-tra](https://github.com/dennis-tra).

## Contributing

Feel free to dive in! [Open an issue](https://github.com/dennis-tra/nebula/issues/new) or submit PRs.

## Support

It would really make my day if you supported this project through [Buy Me A Coffee](https://www.buymeacoffee.com/dennistra).

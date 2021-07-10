// Synchronous database opening
const fs = require('fs');
const Reader = require('@maxmind/geoip2-node').Reader;

const dbBuffer = fs.readFileSync('GeoLite2-Country.mmdb');

// This reader object should be reused across lookups as creation of it is
// expensive.
const reader = Reader.openBuffer(dbBuffer);

response = reader.country('188.63.76.112');

console.log(response.country.isoCode);
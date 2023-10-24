package config

import "fmt"

type Database struct {
	// Whether to skip database interactions
	DryRun bool

	// File path to the JSON output directory
	JSONOut string

	// Determines the host address of the database.
	DatabaseHost string

	// Determines the port of the database.
	DatabasePort int

	// Determines the name of the database that should be used.
	DatabaseName string

	// Determines the password with which we access the database.
	DatabasePassword string

	// Determines the username with which we access the database.
	DatabaseUser string

	// Postgres SSL mode (should be one supported in https://www.postgresql.org/docs/current/libpq-ssl.html)
	DatabaseSSLMode string

	// The cache size to hold agent versions in memory to skip database queries.
	AgentVersionsCacheSize int

	// The cache size to hold protocols in memory to skip database queries.
	ProtocolsCacheSize int

	// The cache size to hold sets of protocols in memory to skip database queries.
	ProtocolsSetCacheSize int
}

// DatabaseSourceName returns the data source name string to be put into the sql.Open method.
func (c *Database) DatabaseSourceName() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseName,
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseSSLMode,
	)
}

package config

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"
)

type Network string

const (
	NetworkIPFS     Network = "IPFS"
	NetworkAmino    Network = "AMINO"
	NetworkFilecoin Network = "FILECOIN"
	NetworkEthereum Network = "ETHEREUM" // TODO: be more specific CL vs EL?
	NetworkKusama   Network = "KUSAMA"
	NetworkPolkadot Network = "POLKADOT"
	NetworkRococo   Network = "ROCOCO"
	NetworkWestend  Network = "WESTEND"
	NetworkArabica  Network = "ARABICA"
	NetworkMocha    Network = "MOCHA"
	NetworkBlockRa  Network = "BLOCKSPACE_RACE"
)

// Root contains general user configuration.
type Root struct {
	// The version string of nebula
	RawVersion string

	// Enables debug logging (equivalent to log level 5)
	Debug bool

	// Specific log level from 0 (least verbose) to 6 (most verbose)
	LogLevel int

	// The time to wait until a dial attempt is aborted.
	DialTimeout time.Duration

	// Determines where the prometheus and pprof hosts should bind to.
	TelemetryHost string

	// Determines the port where prometheus and pprof serve the metrics endpoint.
	TelemetryPort int

	// Contains all configuration parameters for interacting with the database
	Database *Database
}

// Version returns the actual version string which includes VCS information
func (r *Root) Version() string {
	shortCommit := ""
	dirty := false
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				shortCommit = setting.Value[:7]
			case "vcs.modified":
				dirty, _ = strconv.ParseBool(setting.Value)
			}
		}
	}

	versionStr := "v" + r.RawVersion

	if shortCommit != "" {
		versionStr += "+" + shortCommit
	}

	if dirty {
		versionStr += "+dirty"
	}

	return versionStr
}

// String prints the configuration as a json string
func (r *Root) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return fmt.Sprintf("%s", data)
}

type Monitor struct {
	Root *Root

	// How many parallel workers should crawl the network.
	MonitorWorkerCount int
}

// String prints the configuration as a json string
func (m *Monitor) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}

type Resolve struct {
	Root *Root

	FilePathUdgerDB string
	BatchSize       int
}

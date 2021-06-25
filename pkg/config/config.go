package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	Prefix     = "nebula"
	ContextKey = "config"
)

// DefaultConfig the default configuration.
var DefaultConfig = Config{
	BootstrapPeers: []string{}, // see init
	DialTimeout:    10 * time.Second,
	WorkerCount:    500,
	PrometheusPort: 6666,
}

func init() {
	for _, maddr := range dht.DefaultBootstrapPeers {
		DefaultConfig.BootstrapPeers = append(DefaultConfig.BootstrapPeers, maddr.String())
	}
}

// configFile contains the path suffix that's appended to
// an XDG compliant directory to find the settings file.
var configFile = filepath.Join(Prefix, "config.json")

// Config contains general user configuration.
type Config struct {
	// The path where the configuration file is located.
	Path string `json:"-"`

	// Whether the configuration file existed when the tool was started
	Exists bool `json:"-"`

	// The list of multi addresses that will make up the entry points to the network.
	BootstrapPeers []string

	// The time to wait until a dial attempt is aborted.
	DialTimeout time.Duration

	// How many parallel workers should crawl the network.
	WorkerCount int

	// On which port should prometheus serve the metrics endpoint.
	PrometheusPort int
}

// Save saves the peer settings and identity information
// to disk.
func (c *Config) Save() error {
	log.Infoln("Saving configuration file to", c.Path)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(c.Path, data, 0o744); err == nil {
		c.Exists = true
	}
	return err
}

// Apply takes command line arguments and overwrites the respective configurations.
func (c *Config) Apply(ctx *cli.Context) {
	if ctx.IsSet("workers") {
		c.WorkerCount = ctx.Int("workers")
	}
	if ctx.IsSet("dial-timeout") {
		c.DialTimeout = ctx.Duration("dial-timeout")
	}
	if ctx.IsSet("prom-port") {
		c.PrometheusPort = ctx.Int("prom-port")
	}
}

// Apply takes command line arguments and overwrites the respective configurations.
func (c *Config) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return fmt.Sprintf("%s", data)
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		// If no configuration file was given use xdg file.
		var err error
		path, err = xdg.ConfigFile(configFile)
		if err != nil {
			return nil, err
		}
	}
	log.Debugln("Using configuration file at:", path)
	config := DefaultConfig
	config.Path = path
	data, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(data, &config)
		if err != nil {
			return nil, err
		}
		config.Exists = true
		return &config, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	} else {
		return &config, config.Save()
	}
}

func FillContext(c *cli.Context) (*cli.Context, error) {
	conf, err := LoadConfig(c.String("config"))
	if err != nil {
		return c, err
	}

	// Apply command line argument configurations.
	conf.Apply(c)

	// Print full configuration.
	log.Traceln("Configuration (CLI params overwrite file config):\n", conf)

	// Populate the context with the configuration.
	c.Context = context.WithValue(c.Context, ContextKey, conf)
	return c, nil
}

func FromContext(ctx context.Context) (*Config, error) {
	obj := ctx.Value(ContextKey)
	if obj == nil {
		return nil, fmt.Errorf("config not found in context")
	}

	config, ok := obj.(*Config)
	if !ok {
		return nil, fmt.Errorf("config not found in context")
	}

	return config, nil
}

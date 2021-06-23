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
	"github.com/urfave/cli/v2"
)

const (
	Prefix     = "nebula"
	ContextKey = "config"
)

// DefaultConfig the default configuration.
var DefaultConfig = Config{
	BootstrapNodes: nil,
	ConnectTimeout: 10 * time.Second,
	WorkerCount:    500,
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
	BootstrapNodes []string

	// The time to wait until a dial attempt is aborted.
	ConnectTimeout time.Duration

	// How many parallel workers should crawl the network.
	WorkerCount int
}

// Save saves the peer settings and identity information
// to disk.
func (c *Config) Save() error {
	err := save(configFile, c, 0o744)
	if err == nil {
		c.Exists = true
	}
	return err
}

func LoadConfig() (*Config, error) {
	path, err := xdg.ConfigFile(configFile)
	if err != nil {
		return nil, err
	}

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
	conf, err := LoadConfig()
	if err != nil {
		return c, err
	}
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

func save(relPath string, obj interface{}, perm os.FileMode) error {
	path, err := xdg.ConfigFile(relPath)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, perm)
}

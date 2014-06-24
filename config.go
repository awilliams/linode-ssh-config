package main

import (
	"fmt"

	"code.google.com/p/gcfg"
)

// read and parse a config file. if configPath is empty value, defaults to ~/.linode-ssh-config.ini
func loadConfig(configPath string) (*configuration, error) {
	var err error
	if !fileExists(configPath) {
		return nil, fmt.Errorf("no config file found at: %s", configPath)
	}

	var cfg struct {
		Linode configuration
	}
	if err = gcfg.ReadFileInto(&cfg, configPath); err != nil {
		return nil, err
	}
	for _, dg := range cfg.Linode.DisplayGroups {
		cfg.Linode.displayGroupLookup[dg] = empty{}
	}

	return &cfg.Linode, nil
}

type empty struct{}

type configuration struct {
	APIKey             string           `gcfg:"api-key"`       // Linode API key
	DisplayGroups      []string         `gcfg:"display-group"` // Which DisplayGroups to use, if empty, use all
	Running            bool             `gcfg:"running"`       // Only consider running Linodes
	User               string           `gcfg:"user"`          // ssh User
	IdentityFile       string           `gcfg:"identity-file"` // ssh IdentityFile
	displayGroupLookup map[string]empty // used internally for fast displaygroup lookups
}

func (c configuration) filterDisplayGroup(displayGroup string) bool {
	// No display groups matches all
	if len(c.DisplayGroups) == 0 {
		return true
	}
	_, contains := c.displayGroupLookup[displayGroup]
	return contains
}

func (c configuration) filterRunning(isRunning bool) bool {
	if c.Running {
		return isRunning
	}
	return true
}

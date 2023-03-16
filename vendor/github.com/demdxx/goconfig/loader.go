package goconfig

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	env "github.com/caarlos0/env/v6"
	"github.com/gravitational/configure"
	"github.com/hashicorp/hcl"
	defaults "github.com/mcuadros/go-defaults"
)

type configFilepath interface {
	ConfigFilepath() string
}

type config any

// Load data from file
func Load(cfg config) (err error) {
	// Set defaults for config
	defaults.SetDefaults(cfg)

	// parse command line arguments
	if len(os.Args) > 1 {
		if err = configure.ParseCommandLine(cfg, os.Args[1:]); err != nil {
			return err
		}
	}

	// parse config from file
	if configFile, _ := cfg.(configFilepath); configFile != nil {
		if filepath := configFile.ConfigFilepath(); len(filepath) > 0 {
			if err = loadFile(cfg, filepath); err != nil {
				return err
			}
		}
	}

	// parse environment variables
	if err = env.Parse(cfg); err != nil {
		return err
	}

	// parse command line arguments
	if len(os.Args) > 1 {
		err = configure.ParseCommandLine(cfg, os.Args[1:])
	}
	return err
}

// loadFile config from file path
func loadFile(cfg config, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".yml", ".yaml":
		return configure.ParseYAML(data, cfg)
	case ".json":
		return json.Unmarshal(data, cfg)
	case ".hcl":
		var root any
		// For some specific HCL module not always work as expected
		// so this is a hack to fix it
		if err = hcl.Unmarshal(data, &root); err != nil {
			return err
		}
		if data, err = json.Marshal(root); err != nil {
			return err
		}
		// Skip the error because of HCL converts structures into arrays of structs
		_ = json.Unmarshal(data, cfg)
		return hcl.Unmarshal(data, cfg)
	}
	return fmt.Errorf("unsupported config ext: %s", ext)
}

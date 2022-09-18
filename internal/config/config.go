// Package config provides two ways to configure app
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
)

//Config contains variables to configure app
type Config struct {
	Addr            string `env:"SERVER_ADDRESS"     json:"server_address"`
	BaseURL         string `env:"BASE_URL"           json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"  json:"file_storage_path"`
	UserKey         string `env:"USER_KEY" envDefault:"PaSsW0rD" json:"user_key"`
	DatabaseDSN     string `env:"DATABASE_DSN"       json:"database_dsn"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"       json:"enable_https"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"     json:"trusted_subnet"`
	Config          string `env:"CONFIG"             json:"-"`
}

func (c Config) String() string {
	return fmt.Sprintf(
		"Configuration:\n"+
			"  Addr: %s\n"+
			"  BaseURL: %s\n"+
			"  FileStoragePath: %s\n"+
			"  UserKey: %s\n"+
			"  DatabaseDSN: %s\n"+
			"  EnableHTTPS: %t\n", c.Addr, c.BaseURL, c.FileStoragePath, c.UserKey, c.DatabaseDSN, c.EnableHTTPS,
	)
}

//DatabaseDSN scheme: "postgres://username:password@localhost:5432/database_name?sslmode=disable"

//Parse parses command line parameters or sets defaults (if parameters were not set) and then overrides them by environment variables
func (c *Config) Parse() error {
	// priority: config file -> flags -> env

	tempConf := Config{}
	flag.StringVar(&tempConf.Addr, "a", "localhost:8080", "Host to listen on")
	flag.StringVar(&tempConf.BaseURL, "b", "localhost:8080/", "Base address of the resulting shortened URL")
	flag.StringVar(&tempConf.FileStoragePath, "f", "", "Path to the file with shortened URLs")
	flag.StringVar(&tempConf.UserKey, "p", "", "UserKey for encryption cookie")
	flag.StringVar(&tempConf.DatabaseDSN, "d", "", "The line with the address to connect to the database")
	flag.BoolVar(&tempConf.EnableHTTPS, "s", false, "Enable HTTPS mode in web-server")
	flag.StringVar(&tempConf.TrustedSubnet, "t", "192.168.1.0/24", "Trusted subnet")
	flag.StringVar(&tempConf.Config, "config", "", "Config file")
	flag.StringVar(&tempConf.Config, "c", "", "Config file")
	flag.Parse()

	if isFlagPassed("config") || isFlagPassed("c") {
		data, err := os.ReadFile(tempConf.Config)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, c)
		if err != nil {
			return err
		}
	}

	// settings redefinition from flags
	if isFlagPassed("a") || c.Addr == "" {
		c.Addr = tempConf.Addr
	}
	if isFlagPassed("b") || c.BaseURL == "" {
		c.BaseURL = tempConf.BaseURL
	}
	if isFlagPassed("f") || c.FileStoragePath == "" {
		c.FileStoragePath = tempConf.FileStoragePath
	}
	if isFlagPassed("p") || c.UserKey == "" {
		c.UserKey = tempConf.UserKey
	}
	if isFlagPassed("d") || c.DatabaseDSN == "" {
		c.DatabaseDSN = tempConf.DatabaseDSN
	}
	if isFlagPassed("s") {
		c.EnableHTTPS = tempConf.EnableHTTPS
	}
	if isFlagPassed("t") {
		c.TrustedSubnet = tempConf.TrustedSubnet
	}

	// settings redefinition from evn
	err := env.Parse(c)

	if !strings.HasPrefix(c.BaseURL, "http") {
		if c.EnableHTTPS {
			c.BaseURL = "https://" + c.BaseURL
		} else {
			c.BaseURL = "http://" + c.BaseURL
		}
	}

	if !strings.HasSuffix(c.BaseURL, "/") {
		c.BaseURL = c.BaseURL + "/"
	}

	fmt.Println(c)

	return err
}

// isFlagPassed checks whether the flag was set in CLI
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr			string		`env:"SERVER_ADDRESS"`
	BaseURL			string		`env:"BASE_URL"`
	FileStoragePath	string		`env:"FILE_STORAGE_PATH"`
}

func (c* Config)Parse() error {
	flag.StringVar(&c.Addr, "a", "localhost:8080", "host to listen on")
	flag.StringVar(&c.BaseURL,"b", "http://localhost:8080", "Base address of the resulting shortened URL")
	flag.StringVar(&c.FileStoragePath,"f", "", "Path to the file with shortened URLs")
	flag.Parse()

	//settings redefinition, if evn variables is used
	return env.Parse(c)
}
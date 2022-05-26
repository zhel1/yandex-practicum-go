package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"strings"
)

type Config struct {
	Addr			string		`env:"SERVER_ADDRESS"`
	BaseURL			string		`env:"BASE_URL"`
	FileStoragePath	string		`env:"FILE_STORAGE_PATH"`
	UserKey			string		`env:"USER_KEY" envDefault:"PaSsW0rD"`
}

func (c* Config)Parse() error {
	flag.StringVar(&c.Addr, "a", "localhost:8080", "Host to listen on")
	flag.StringVar(&c.BaseURL,"b", "http://localhost:8080/", "Base address of the resulting shortened URL")
	flag.StringVar(&c.FileStoragePath,"f", "C:\\Users\\Денис\\go\\src\\yandex-practicum-go\\db.txt", "Path to the file with shortened URLs")
	flag.StringVar(&c.UserKey,"p", "", "UserKey for encryption cookie")
	flag.Parse()

	//settings redefinition, if evn variables is used
	err := env.Parse(c)

	if !strings.HasSuffix(c.BaseURL, "/") {
		c.BaseURL = c.BaseURL + "/"
	}

	return err
}
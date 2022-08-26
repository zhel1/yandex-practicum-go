package main

import (
	"fmt"
	"github.com/zhel1/yandex-practicum-go/internal/app"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildData() {
	switch buildVersion {
	case "":
		fmt.Printf("Build version: %s\n", "N/A")
	default:
		fmt.Printf("Build version: %s\n", buildVersion)
	}
	switch buildDate {
	case "":
		fmt.Printf("Build date: %s\n", "N/A")
	default:
		fmt.Printf("Build date: %s\n", buildDate)
	}
	switch buildCommit {
	case "":
		fmt.Printf("Build commit: %s\n", "N/A")
	default:
		fmt.Printf("Build commit: %s\n", buildCommit)
	}
}

func main() {
	printBuildData()
	app.Run()
}

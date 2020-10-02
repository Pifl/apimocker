package main

import (
	"apimocker/api"
	"fmt"

	"github.com/BurntSushi/toml"
)

type tomlConfig struct {
	Admin string
	API   apiInfo
}

type apiInfo struct {
	Port string
}

func main() {
	var config tomlConfig
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println(err)
		return
	}
	api.Start(config.API.Port)
}

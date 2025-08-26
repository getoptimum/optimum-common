package main

import (
	"fmt"

	cfg "github.com/getoptimum/optimum-common/config"
)

type AppConfig struct {
	Name  string `yaml:"name"`
	Port  int    `yaml:"port"`
	Debug bool   `yaml:"debug"`
}

func main() {
	cfgStruct := AppConfig{Port: 8080}
	defs := []cfg.FlagDef{
		{Name: "name", Usage: "application name", Value: &cfgStruct.Name},
		{Name: "port", Usage: "port to listen on", Value: &cfgStruct.Port},
		{Name: "debug", Usage: "enable debug logging", Value: &cfgStruct.Debug},
	}
	if err := cfg.Load(&cfgStruct, defs); err != nil {
		panic(err)
	}
	fmt.Printf("effective config: %+v\n", cfgStruct)
}

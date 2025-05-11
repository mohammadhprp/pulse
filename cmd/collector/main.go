package main

import (
	"log"

	"github.com/mohammadhptp/pulse/internal/collector"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	collector.Run()
}

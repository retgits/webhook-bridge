package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"
)

var ()

func main() {
	// Get configuration parameters from config
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/webhookbridge/")
	viper.AddConfigPath("$HOME/.webhookbridge")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error reading config file: %s", err))
	}

	// Set the default level
	loglevel, err := zerolog.ParseLevel(viper.GetString("loglevel"))
	if err != nil {
		panic(fmt.Errorf("fatal error reading log level: %s", err))
	}
	zerolog.SetGlobalLevel(loglevel)

	// Enable ConsoleWriter only for runmode debug
	if loglevel == zerolog.DebugLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Create a new server
	srv := NewServer()

	// Create a channel to wait for quit signals
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	// Create a go routine that will wait for a signal interrupt
	// to gracefully shutdown the server
	go func() {
		<-quit
		log.Info().Msg("Received os.Interrupt signal")
		srv.Stop()
		os.Exit(0)
	}()

	// Start the server
	srv.Start()
}

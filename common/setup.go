package common

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	c Config
)

// Config contains the common configuration data for the server.
type Config struct {
	LogLevel string `required:"true"`
}

// HandleSetup takes care of initializing the logger and a few other components.
// It's done in the common package so that all agents making use of this function
// can do the exact same thing.
func HandleSetup() {
	// Create a config struct and read in the environent variables required
	err := envconfig.Process("", &c)
	if err != nil {
		panic(fmt.Errorf("fatal error reading environment variables: %s", err.Error()))
	}

	// Set the default level
	loglevel, err := zerolog.ParseLevel(c.LogLevel)
	if err != nil {
		panic(fmt.Errorf("fatal error reading log level: %s", err))
	}
	zerolog.SetGlobalLevel(loglevel)

	// Enable ConsoleWriter only for runmode debug
	if loglevel == zerolog.DebugLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

package main

import (
	"os"
	"os/signal"

	"github.com/retgits/webhook-bridge/common"
	"github.com/retgits/webhook-bridge/jenkins"
	"github.com/rs/zerolog/log"
)

func main() {
	// Handle common startup processes
	common.HandleSetup()

	// Create a new server
	srv, err := jenkins.Register()
	if err != nil {
		log.Fatal().Msgf("fatal error while creating server: %s", err.Error())
	}

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
	done := srv.Start()
	log.Info().Msg("Started server successfully and waiting for messages...")
	<-done
}

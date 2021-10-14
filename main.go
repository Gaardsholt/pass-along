package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Gaardsholt/pass-along/api"
	"github.com/Gaardsholt/pass-along/config"
	"github.com/rs/zerolog/log"
)

func init() {
	config.LoadConfig()
}

func main() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	internalServer, externalServer := api.StartServer()

	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Debug().Msg("Got SIGINT...")
	case syscall.SIGTERM:
		log.Debug().Msg("Got SIGTERM...")
	}

	log.Info().Msg("The service is shutting down...")
	externalServer.Shutdown(context.Background())
	internalServer.Shutdown(context.Background())
}

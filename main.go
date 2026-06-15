package main

import (
	"context"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Gaardsholt/pass-along/api"
	"github.com/Gaardsholt/pass-along/config"
	"github.com/rs/zerolog/log"
)

func init() {
	config.LoadConfig()
}

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ongoingCtx, cancelOngoing := context.WithCancel(context.Background())
	defer cancelOngoing()

	var isShuttingDown atomic.Bool
	internalServer, externalServer, closeStore := api.StartServer(ongoingCtx, &isShuttingDown)

	<-rootCtx.Done()
	stop()

	log.Info().Msg("The service is shutting down")
	isShuttingDown.Store(true)

	readinessDrainDelay := config.Config.GetReadinessDrainDelay()
	log.Info().Dur("delay", readinessDrainDelay).Msg("Waiting for readiness change to propagate")
	time.Sleep(readinessDrainDelay)

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), config.Config.GetHTTPShutdownTimeout())
	defer cancelShutdown()

	if err := externalServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Error shutting down external server")
	}
	if err := internalServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Error shutting down internal server")
	}

	cancelOngoing()
	if shutdownCtx.Err() != nil {
		hardDelay := config.Config.GetShutdownHardDelay()
		log.Warn().Dur("delay", hardDelay).Msg("Shutdown timeout reached; allowing request contexts to observe cancellation")
		time.Sleep(hardDelay)
	}

	if err := closeStore(); err != nil {
		log.Error().Err(err).Msg("Error closing datastore")
	}

	log.Info().Msg("The service shut down gracefully")
}

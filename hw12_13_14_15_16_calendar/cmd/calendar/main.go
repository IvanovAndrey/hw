package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/consts"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/app"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/rs/zerolog/log"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/Users/andrey.ivanov/repos/Education/hw/hw12_13_14_15_16_calendar/configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := LoadConfig(configFile)
	if err != nil {
		panic(err)
	}

	logg := logger.New(consts.AppName, release, cfg.Logger.Level)

	storage := memorystorage.NewLocalStorage(logg.WithModule("localStorage"))
	calendar := app.New(logg.WithModule("dbStorage"), storage)

	server := internalhttp.NewServer(logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		cancel()
		log.Fatal().Msgf("failed to start http server: %s", err)
	}
}

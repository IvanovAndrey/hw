package main

import (
	"context"
	"flag"
	"os/signal"
	"sync"
	"syscall"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/configuration"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/consts"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/app"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/server/http"
	storage2 "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./../../configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := configuration.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}

	logg := logger.NewLogger(consts.AppName, release, cfg.Logger.Level)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	var storage storage2.Storage
	if cfg.System.Database.Enable {
		dbCfg := sqlstorage.DBConfig{
			Host:     cfg.System.Database.Host,
			Port:     cfg.System.Database.Port,
			User:     cfg.System.Database.User,
			Password: cfg.System.Database.Password,
			DBName:   cfg.System.Database.DBName,
			SSLMode:  cfg.System.Database.SSLMode,
		}
		storage, err = sqlstorage.NewStorage(ctx, dbCfg, logg.WithModule("sqlStorage"))
		if err != nil {
			logg.Fatal("failed to connect to db")
		}
	} else {
		storage = memorystorage.NewLocalStorage(logg.WithModule("localStorage"))
	}
	calendar := app.New(logg.WithModule("app"), storage)

	server := internalhttp.NewServer(cfg, *logg, calendar)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(); err != nil {
		cancel()
		logg.Fatal("failed to start http server:" + err.Error())
	}
	wg.Wait()
}

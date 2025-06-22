package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/cmd"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/configuration"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/consts"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/rmq"
	storageInterface "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/scheduler_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	cfg, err := configuration.LoadSchedulerConfig(configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logg := logger.NewLogger(consts.SchedulerAppName, cmd.Release, cfg.Logger.Level)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		logg.Info("shutting down scheduler")
		cancel()
	}()

	var storage storageInterface.Storage
	if cfg.System.Database.Enable {
		storage, err = sqlstorage.NewStorage(ctx, &cfg.System.Database, logg.WithModule("sqlStorage"))
		if err != nil {
			logg.Fatal("failed to connect to db")
		}
	} else {
		storage = memorystorage.NewLocalStorage(logg.WithModule("localStorage"))
	}

	rmqClient, err := rmq.NewRMQClient(cfg.RabbitMQ.URI, cfg.RabbitMQ.Queue)
	if err != nil {
		logg.Fatal(fmt.Sprintf("failed to init rmq: %v", err))
	}
	defer rmqClient.Close()

	ticker := time.NewTicker(cfg.Scheduler.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			events, err := storage.EventsToNotify(ctx)
			if err != nil {
				logg.Error(fmt.Sprintf("query notify events: %v", err))
				continue
			}
			for _, ev := range events {
				eventTime, err := time.Parse(time.RFC3339, ev.Date)
				if err != nil {
					logg.Error(fmt.Sprintf("invalid event date: %v", err))
					continue
				}
				note := rmq.Notification{
					EventID:  ev.ID,
					Title:    ev.Title,
					DateTime: eventTime,
					UserID:   ev.User,
				}
				err = rmqClient.PublishNotification(ctx, note)
				if err != nil {
					logg.Error(fmt.Sprintf("publish notification: %v", err))
				}
			}

			err = storage.DeleteOldEvents(ctx, time.Now().AddDate(-1, 0, 0))
			if err != nil {
				logg.Error(fmt.Sprintf("cleanup old events: %v", err))
			}
		}
	}
}

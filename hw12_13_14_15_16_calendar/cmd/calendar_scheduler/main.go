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

	var storage storageInterface.Storage

	if cfg.System.Database.Enable {
		storage, err = sqlstorage.NewStorage(ctx, &cfg.System.Database, logg.WithModule("sqlStorage"))
		if err != nil {
			logg.Fatal(fmt.Sprintf("failed to connect to db: %v", err))
		}
	} else {
		storage = memorystorage.NewLocalStorage(logg.WithModule("localStorage"))
	}

	go func() {
		<-stop
		logg.Info("shutting down scheduler")
		cancel()
	}()

	if err := runScheduler(ctx, cfg, storage, logg); err != nil {
		logg.Fatal(fmt.Sprintf("scheduler stopped with error: %v", err))
	}
}

func runScheduler(
	ctx context.Context,
	cfg *configuration.SchedulerConfig,
	storage storageInterface.Storage,
	logg *logger.Logger,
) error {
	var err error
	rmqClient, err := rmq.NewRMQClient(cfg.RabbitMQ.URI, cfg.RabbitMQ.Queue)
	if err != nil {
		return fmt.Errorf("failed to init rmq: %w", err)
	}
	defer func() {
		if err := rmqClient.Close(); err != nil {
			logg.Error("failed to close rmq client")
		}
	}()

	ticker := time.NewTicker(cfg.Scheduler.Interval)
	logg.Info(fmt.Sprintf("Scheduler was started with interval %v", cfg.Scheduler.Interval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			events, err := storage.EventsToNotify(ctx)
			if err != nil {
				logg.Error(fmt.Sprintf("query notify events: %v", err))
				continue
			}
			if len(events) == 0 {
				logg.Debug("no events to notify")
			} else {
				for _, ev := range events {
					note := rmq.Notification{
						EventID:  ev.ID,
						Title:    ev.Title,
						DateTime: ev.Date,
						UserID:   ev.User,
					}

					if err := rmqClient.PublishNotification(ctx, note); err != nil {
						logg.Error(fmt.Sprintf("publish notification: %v", err))
					}
				}
			}
			if err := storage.DeleteOldEvents(ctx, time.Now().AddDate(-1, 0, 0)); err != nil {
				logg.Error(fmt.Sprintf("cleanup old events: %v", err))
			}
		}
	}
}

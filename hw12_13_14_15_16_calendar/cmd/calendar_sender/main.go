package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/cmd"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/configuration"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/rmq"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/sender_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	cfg, err := configuration.LoadSenderConfig(configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logg := logger.NewLogger("calendar_sender", cmd.Release, cfg.Logger.Level)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		logg.Info("shutting down sender")
		cancel()
	}()

	rmqClient, err := rmq.NewRMQClient(cfg.RabbitMQ.URI, cfg.RabbitMQ.Queue)
	if err != nil {
		logg.Fatal(fmt.Sprintf("failed to init rmq: %v", err))
	}
	defer func(rmqClient *rmq.RMQClient) {
		err := rmqClient.Close()
		if err != nil {
			logg.Fatal(fmt.Sprintf("failed to close rmq client: %v", err))
		}
	}(rmqClient)

	err = rmqClient.ConsumeNotifications(ctx, func(note rmq.Notification) {
		logg.Info(fmt.Sprintf("[NOTIFY] EventID: %s, Title: %s, DateTime: %s, UserID: %s",
			note.EventID, note.Title, note.DateTime.Format("2006-01-02 15:04:05"), note.UserID))
	})
	if err != nil {
		logg.Fatal(fmt.Sprintf("failed to consume notifications: %v", err))
	}
}

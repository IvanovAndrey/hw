package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		log.Fatal("Usage: go-telnet [--timeout=10s] host port")
	}
	host := args[0]
	port := args[1]

	client := NewTelnetClient(
		net.JoinHostPort(host, port),
		*timeout,
		os.Stdin,
		os.Stdout,
	)

	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	defer func(client TelnetClient) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(client)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		if err := client.Send(); err != nil {
			log.Println("...Connection was closed by peer")
			os.Exit(0)
		}
		log.Println("...EOF")
		os.Exit(0)
	}()

	go func() {
		if err := client.Receive(); err != nil {
			log.Println("...Connection was closed by peer")
			os.Exit(0)
		}
	}()

	<-sigCh
}

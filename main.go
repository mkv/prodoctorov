package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"prodoctorov/internal/service"
)

func main() {
	configFileName := flag.String("config", "cfg/config.yaml", "Name of config file in Yaml format")
	flag.Parse()

	log.Printf("Server starting with config file: %s", *configFileName)

	s, err := service.NewService(*configFileName)
	if err != nil {
		log.Printf("Failed to initialize service: %v", err)

		return
	}

	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt)

	defer func() {
		signal.Stop(interruptSignal)
		close(interruptSignal)
	}()

	if err := s.Run(interruptSignal); err != nil {
		log.Printf("Run error: %v", err)
	}
}

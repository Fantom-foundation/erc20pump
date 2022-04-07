// Package erc20pump implements the application server entry.
package main

import (
	"erc20pump/internal/scanner"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// main provides application entry point
func main() {
	// make the scanner
	s, err := scanner.New(config())
	if err != nil {
		return
	}

	captureTerminate(s)

	// start the scanner
	s.Run()
	log.Println("done")
}

// captureTerminate setups terminate signals observation.
func captureTerminate(s *scanner.Service) {
	// make the signal consumer
	ts := make(chan os.Signal, 1)
	signal.Notify(ts, syscall.SIGINT, syscall.SIGTERM)

	// start monitoring
	go func() {
		<-ts
		s.Stop()
	}()
}

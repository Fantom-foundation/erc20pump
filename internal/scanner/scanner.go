// Package scanner performs the scanning task.
package scanner

import (
	"erc20pump/internal/cfg"
	"erc20pump/internal/scanner/cache"
	"erc20pump/internal/scanner/rpc"
	"sync"
)

// Service represents the scanner manager.
type Service struct {
	wg *sync.WaitGroup
	lp *logPuller
	lc *logCollector
	se *sender
}

// New creates a new scanner service based on provided configuration.
func New(c *cfg.Config) (*Service, error) {
	wg := new(sync.WaitGroup)

	// create blockchain node adapter
	ada, err := rpc.New(c)
	if err != nil {
		return nil, err
	}

	// create cache
	cch := cache.New()

	// make sub-services
	lp := newPuller(c, ada, cch)
	lc := newCollector(c, lp.output, ada, cch)
	se := newSender(c, lc.output)

	// build the manager
	return &Service{
		wg: wg,
		lp: lp,
		lc: lc,
		se: se,
	}, nil
}

// Run the scanner service.
func (s *Service) Run() {
	// start all needed threads
	s.se.run(s.wg)
	s.lc.run(s.wg)
	s.lp.run(s.wg)

	// wait until all the threads terminate
	s.wg.Wait()
}

// Stop the scanner service by signaling sub-services to terminate.
func (s *Service) Stop() {
	s.lp.stop()
	s.lc.stop()
	s.se.stop()
}

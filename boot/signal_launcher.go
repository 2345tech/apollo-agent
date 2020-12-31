package boot

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

type DataChan chan struct{}

type SignalLauncher struct {
	booted bool
	agent  *Agent
	signal chan os.Signal
}

func NewSignal() *SignalLauncher {
	return &SignalLauncher{
		booted: false,
		signal: make(chan os.Signal),
	}
}

func (s *SignalLauncher) Init(agent *Agent) error {
	s.agent = agent
	agent.SignalL = s
	return nil
}

func (s *SignalLauncher) Run() error {
	if s.booted {
		return nil
	}
	go s.watchOsSignal()
	s.booted = true
	return nil
}

func (s *SignalLauncher) Stop() {
}

func (s *SignalLauncher) Shutdown() {
	close(s.signal)
	s.booted = false
	log.Println("[INFO] SignalLauncher Shutdown")
}

func (s *SignalLauncher) watchOsSignal() {
	signal.Notify(s.signal, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		switch <-s.signal {
		case syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT:
			log.Println("[INFO] agent get close signal...")
			s.agent.SigBus.StopS <- struct{}{}
		}
	}
}

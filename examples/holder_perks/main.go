package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aegir-tactics/watchmen"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.Info("server starting...")

	s, err := NewServer(logger)
	if err != nil {
		logger.Fatal(err)
	}
	wm, err := watchmen.New(logger, s)
	if err != nil {
		logger.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go func() {
		<-sigs
		close(sigs)
		logger.Info("gracefully shutting down server")
		cancelFunc()
	}()

	logger.Info("server started")
	if err := wm.Start(ctx); err != nil {
		logger.Fatal(err)
	}
}

// Server ...
type Server struct {
	logger *logrus.Entry

	holders  map[string]struct{}
	assetIDs map[uint64]struct{}
}

// NewServer ...
func NewServer(logger *logrus.Logger) (*Server, error) {
	s := &Server{}
	s.logger = logger.WithField("component", "holder_perks")
	s.holders = map[string]struct{}{
		"holderaddress1": struct{}{},
		"holderaddress2": struct{}{},
	}
	s.assetIDs = map[uint64]struct{}{
		12345: struct{}{},
		12346: struct{}{},
	}

	return s, nil
}

// Dispatch ...
func (s *Server) Dispatch(ctx context.Context, b watchmen.Block) error {
	addresses := map[string]struct{}{}

	// Check if address is of interest
	for _, txn := range b.Txns {
		if _, ok := s.assetIDs[txn.AssetID]; ok {
			addresses[txn.Sender] = struct{}{}
			addresses[txn.Receiver] = struct{}{}
			continue
		}
		if _, ok := s.holders[txn.Sender]; ok {
			addresses[txn.Sender] = struct{}{}
		}
		if _, ok := s.holders[txn.Receiver]; ok {
			addresses[txn.Receiver] = struct{}{}
		}
	}

	// Refresh Address Data
	for address := range addresses {
		if err := s.Refresh(address); err != nil {
			return fmt.Errorf("dispatch: %s", err)
		}
	}

	return nil
}

// Refresh ...
func (s *Server) Refresh(address string) error {
	s.logger.Infof("refresh: %q\n", address)

	return nil
}

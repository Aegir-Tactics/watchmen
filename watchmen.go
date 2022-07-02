package watchmen

import (
	"context"
	"fmt"
	"sync"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/sirupsen/logrus"
)

const (
	MainnetAlgoNode = "https://mainnet-api.algonode.cloud"
	TestnetAlgoNode = "https://testnet-api.algonode.cloud"
)

// Dispatcher ...
type Dispatcher interface {
	Dispatch(context.Context, Block) error
}

// Watcher ...
type Watcher struct {
	Config

	logger      *logrus.Entry
	ac          *algod.Client
	dispatchers []Dispatcher
}

// New ...
func New(logger *logrus.Logger, dispatchers ...Dispatcher) (*Watcher, error) {
	w := &Watcher{}
	w.Config = NewConfig()
	w.logger = logger.WithField("component", "watchmen")
	for _, d := range dispatchers {
		w.dispatchers = append(w.dispatchers, d)
	}

	algodURL := TestnetAlgoNode
	if w.MainnetEnabled {
		algodURL = MainnetAlgoNode
	}

	ac, err := algod.MakeClient(algodURL, "")
	if err != nil {
		return nil, fmt.Errorf("new: %s", err)
	}
	w.ac = ac

	return w, nil
}

// Block ...
type Block struct {
	Round uint64

	Txns []Txn
}

// Txn ...
type Txn struct {
	Sender          string
	Receiver        string
	ApplicationID   uint64
	ApplicationArgs [][]byte

	Amount  uint64
	AssetID uint64
}

// Start ...
func (w *Watcher) Start(ctx context.Context) error {
	status, err := w.ac.Status().Do(ctx)
	if err != nil {
		return fmt.Errorf("start: %s", err)
	}
	var wg sync.WaitGroup

	for {
		block, err := w.ac.Block(status.LastRound).Do(ctx)
		if err != nil {
			return fmt.Errorf("start: %s", err)
		}

		b := Block{
			Round: uint64(block.Round),
			Txns:  []Txn{},
		}
		if b.Round%10 == 0 {
			w.logger.Info("ROUND:", b.Round)
		}
		for _, ps := range block.Payset {
			var txn Txn
			txn.AssetID = uint64(ps.Txn.XferAsset)
			txn.Amount = uint64(ps.Txn.Amount)
			if txn.Amount == 0 {
				txn.Amount = ps.Txn.AssetAmount
			}
			txn.Sender = ps.Txn.Sender.String()
			txn.Receiver = ps.Txn.Receiver.String()
			if !ps.Txn.AssetSender.IsZero() {
				txn.Sender = ps.Txn.AssetSender.String()
			}
			if !ps.Txn.AssetReceiver.IsZero() {
				txn.Receiver = ps.Txn.AssetReceiver.String()
			}

			txn.ApplicationID = uint64(ps.Txn.ApplicationID)
			txn.ApplicationArgs = ps.Txn.ApplicationArgs
			b.Txns = append(b.Txns, txn)
		}

		for _, dispatcher := range w.dispatchers {
			wg.Add(1)
			go func(c Block) {
				if err := dispatcher.Dispatch(ctx, c); err != nil {
					w.logger.Errorf("start: %s", err)
				}
				wg.Done()
			}(b)
		}

		wg.Wait()

		status, err = w.ac.StatusAfterBlock(b.Round).Do(ctx)
		if err != nil {
			return fmt.Errorf("start: %s", err)
		}
	}

	return nil
}

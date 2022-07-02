# watchmen

Watchmen is a component which watches an Algorand Node for new blocks and sends
the information to components which register for block information. Each
registered component is responsible for the logic of handling what to do with
the information.

## How to use

To utilize the watchmen component a new component needs to be created which
implements the `Dispatch(context.Context, watchmen.Block) error` interface.

```go
// Dispatch ...
func (s *Server) Dispatch(ctx context.Context, b watchmen.Block) error {
  addresses := map[string]struct{}{}

  // loop through all transactions in a finalized block
  for _, txn := range b.Txns {
    // Check if asset is of interest
    if _, ok := s.assetIDs[txn.AssetID]; ok {
      addresses[txn.Sender] = struct{}{}
      addresses[txn.Receiver] = struct{}{}
      continue
    }

    // Check if sender address is of interest
    if _, ok := s.holders[txn.Sender]; ok {
      addresses[txn.Sender] = struct{}{}
    }
    // Check if receiver address is of interest
    if _, ok := s.holders[txn.Receiver]; ok {
      addresses[txn.Receiver] = struct{}{}
    }
  }

  // Loop through addresses that need to be updated
  for address := range addresses {
    // Do some sort of action to refresh data for a specific address
    if err := s.Handle(address); err != nil {
      return fmt.Errorf("dispatch: %s", err)
    }
  }

  return nil
}
```

After the component is made which implements the `Dispatch` method. It must be
registered with the watchmen.

```go
// Create New Server instance which implements Dispatch method
s, err := NewServer(logger)
if err != nil {
  logger.Fatal(err)
}

// Add the server instance to the creation of watchmen
wm, err := watchmen.New(logger, s)
if err != nil {
  logger.Fatal(err)
}
```

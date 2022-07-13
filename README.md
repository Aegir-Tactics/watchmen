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

## Use Cases

### HOLDER ROLE STATUS IN DISCORD
In Aegir Tactics discord people can register their wallet to get a `Mythic Holder` role which allows them to enter giveaways and win prizes. But if that user then sells or removes their assets they should no longer have the role. Using watchmen allows us to watch for these type of transactions and update their roles accordingly.

### ARBITRAGE BOTS
Using watchmen to watch addresses of liquidity pools can allow arbitrage bots to execute tasks which capitalize on inbalances created by funds entering or leaving the pool.

### REDUCED INDEXER NODE WITHOUT SETTING UP A FULL ARCHIVAL NODE
Projects often would like to run their own indexer nodes, but the cost and size is quite a challenge. Using watchmen projects can build their own use case specific indexer node which only holds and updates information that the project cares about. This leads to smaller storage requirements and lower costs. 

### AEGIR TACTICS API
Aegir Tactics uses watchmen to power our api which tracks assets which players own. The api is used by our main game, discord mini-game, quest system, and discord holder role management. Using watchmen allows us to have full control over our game state without relying solely on 3rd party services.

Aegir Tactics Card Game
![unknown](https://user-images.githubusercontent.com/5757420/178617973-ea18bb3d-195e-4a97-b91f-bc87c0b318f8.png)

Aegir Dungeons (Discord Mini-Game)
![unknown](https://user-images.githubusercontent.com/5757420/178618016-5b7b8ae3-abfd-434e-a464-b3fd8f527b5d.png)

Mythic Quests
![unknown](https://user-images.githubusercontent.com/5757420/178618055-f2d24137-0b87-47f5-80f9-14b6f0f91bde.png)

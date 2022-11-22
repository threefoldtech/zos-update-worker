# zos-update-worker

A worker to get the version set on the chain with the substrate-client with a specific interval (for example: 10 mins) for mainnet, testnet, and qanet

## How to use

- build `make build`
- Run cmd

```bash
./bin/zos-update-worker 
```
  
- you can run the command with:

```bash
./bin/zos-update-worker --src=tf-autobuilder --dst=tf-zos --interval=10 --main-url=wss://tfchain.grid.tf/ws --main-url=wss://tfchain.grid.tf/ws --test-url=wss://tfchain.test.grid.tf/ws --test-url=wss://tfchain.test.grid.tf/ws --qa-url=wss://tfchain.qa.grid.tf/ws --qa-url=wss://tfchain.qa.grid.tf/ws
```

## Test

```bash
make test
```

## Coverage

```bash
make coverage
```

## Substrate URLs

```go
SUBSTRATE_URLS := map[string][]string{
 "qa":         {"wss://tfchain.qa.grid.tf/ws"},
 "testing":    {"wss://tfchain.test.grid.tf/ws"},
 "production": {"wss://tfchain.grid.tf/ws"},
}
```

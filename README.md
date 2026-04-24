# redis-go

A small Redis-compatible server written in Go.

This v1 MVP supports RESP2 clients and an in-memory string store. It is intended for learning and local experimentation, not production use.

## Supported commands

- `PING [message]`
- `ECHO message`
- `SET key value`
- `GET key`
- `DEL key [key ...]`
- `EXISTS key [key ...]`

## Run

```sh
go run .
```

The server listens on `:6379`.

In another terminal:

```sh
redis-cli PING
redis-cli SET name Ahmed
redis-cli GET name
redis-cli EXISTS name missing
redis-cli DEL name
```

## Test

```sh
go test ./...
```

## Notes

- Data is stored in memory only and is lost when the server stops.
- Persistence, TTLs, transactions, and advanced Redis data types are outside the current MVP scope.

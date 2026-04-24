# redis-go

A small Redis-compatible server written in Go.

This v2 MVP supports RESP2 clients, an in-memory string store, key expiration, and common string operations. It is intended for learning and local experimentation, not production use.

## Supported commands

- `PING [message]`
- `ECHO message`
- `SET key value`
- `SET key value EX seconds`
- `SET key value PX milliseconds`
- `GET key`
- `MSET key value [key value ...]`
- `MGET key [key ...]`
- `INCR key`
- `DECR key`
- `DEL key [key ...]`
- `EXISTS key [key ...]`
- `EXPIRE key seconds`
- `TTL key`
- `PERSIST key`

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
redis-cli MSET city Hyderabad country India
redis-cli MGET name city missing
redis-cli INCR visits
redis-cli DECR visits
redis-cli SET session active EX 10
redis-cli TTL session
redis-cli PERSIST session
redis-cli EXISTS name missing
redis-cli DEL name
```

## Test

```sh
go test ./...
```

## Notes

- Data is stored in memory only and is lost when the server stops.
- Transactions, persistence, clustering, and non-string Redis data types are outside the current MVP scope.

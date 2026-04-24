package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func handleCommand(value Value, store *Store) Value {
	cmd, args, ok := parseCommand(value)
	if !ok {
		return protocolError("ERR expected array command")
	}

	switch cmd {
	case "PING":
		return pingCommand(args)
	case "ECHO":
		return echoCommand(args)
	case "SET":
		return setCommand(args, store)
	case "GET":
		return getCommand(args, store)
	case "DEL":
		return delCommand(args, store)
	case "EXISTS":
		return existsCommand(args, store)
	case "EXPIRE":
		return expireCommand(args, store)
	case "TTL":
		return ttlCommand(args, store)
	case "PERSIST":
		return persistCommand(args, store)
	default:
		return protocolError(fmt.Sprintf("ERR unknown command '%s'", strings.ToLower(cmd)))
	}
}

func parseCommand(value Value) (string, []string, bool) {
	if value.typ != typeArray || len(value.array) == 0 {
		return "", nil, false
	}

	parts := make([]string, 0, len(value.array))
	for _, item := range value.array {
		part, ok := valueToString(item)
		if !ok {
			return "", nil, false
		}
		parts = append(parts, part)
	}

	cmd := strings.ToUpper(parts[0])
	return cmd, parts[1:], true
}

func valueToString(value Value) (string, bool) {
	switch value.typ {
	case typeBulk:
		return value.bulk, true
	case typeString:
		return value.str, true
	default:
		return "", false
	}
}

func pingCommand(args []string) Value {
	switch len(args) {
	case 0:
		return Value{typ: typeString, str: "PONG"}
	case 1:
		return Value{typ: typeBulk, bulk: args[0]}
	default:
		return wrongArity("ping")
	}
}

func echoCommand(args []string) Value {
	if len(args) != 1 {
		return wrongArity("echo")
	}
	return Value{typ: typeBulk, bulk: args[0]}
}

func setCommand(args []string, store *Store) Value {
	if len(args) != 2 && len(args) != 4 {
		return wrongArity("set")
	}

	ttl, ok := setTTL(args)
	if !ok {
		return protocolError("ERR invalid expire time in 'set' command")
	}

	store.SetWithTTL(args[0], args[1], ttl)
	return Value{typ: typeString, str: "OK"}
}

func setTTL(args []string) (time.Duration, bool) {
	if len(args) == 2 {
		return 0, true
	}

	duration, ok := parseTTL(args[2], args[3])
	if !ok {
		return 0, false
	}
	return duration, true
}

func getCommand(args []string, store *Store) Value {
	if len(args) != 1 {
		return wrongArity("get")
	}

	value, ok := store.Get(args[0])
	if !ok {
		return Value{typ: typeNull}
	}
	return Value{typ: typeBulk, bulk: value}
}

func delCommand(args []string, store *Store) Value {
	if len(args) == 0 {
		return wrongArity("del")
	}
	return Value{typ: typeInteger, num: store.Delete(args...)}
}

func existsCommand(args []string, store *Store) Value {
	if len(args) == 0 {
		return wrongArity("exists")
	}
	return Value{typ: typeInteger, num: store.Exists(args...)}
}

func expireCommand(args []string, store *Store) Value {
	if len(args) != 2 {
		return wrongArity("expire")
	}

	ttl, ok := parseTTL("EX", args[1])
	if !ok {
		return protocolError("ERR invalid expire time in 'expire' command")
	}
	if !store.Expire(args[0], ttl) {
		return Value{typ: typeInteger, num: 0}
	}
	return Value{typ: typeInteger, num: 1}
}

func ttlCommand(args []string, store *Store) Value {
	if len(args) != 1 {
		return wrongArity("ttl")
	}
	return Value{typ: typeInteger, num: store.TTL(args[0])}
}

func persistCommand(args []string, store *Store) Value {
	if len(args) != 1 {
		return wrongArity("persist")
	}
	if !store.Persist(args[0]) {
		return Value{typ: typeInteger, num: 0}
	}
	return Value{typ: typeInteger, num: 1}
}

func parseTTL(unit, raw string) (time.Duration, bool) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, false
	}

	switch strings.ToUpper(unit) {
	case "EX":
		return time.Duration(value) * time.Second, true
	case "PX":
		return time.Duration(value) * time.Millisecond, true
	default:
		return 0, false
	}
}

func wrongArity(command string) Value {
	return protocolError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", command))
}

func protocolError(message string) Value {
	return Value{typ: typeError, str: message}
}

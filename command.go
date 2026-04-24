package main

import (
	"fmt"
	"strings"
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
	if len(args) != 2 {
		return wrongArity("set")
	}
	store.Set(args[0], args[1])
	return Value{typ: typeString, str: "OK"}
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

func wrongArity(command string) Value {
	return protocolError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", command))
}

func protocolError(message string) Value {
	return Value{typ: typeError, str: message}
}

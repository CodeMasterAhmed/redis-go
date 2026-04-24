package main

import (
	"testing"
	"time"
)

func TestHandleCommandPing(t *testing.T) {
	store := NewStore()

	got := handleCommand(commandValue("PING"), store)
	if got.typ != typeString || got.str != "PONG" {
		t.Fatalf("PING = %#v, want PONG simple string", got)
	}

	got = handleCommand(commandValue("PING", "hello"), store)
	if got.typ != typeBulk || got.bulk != "hello" {
		t.Fatalf("PING hello = %#v, want hello bulk string", got)
	}
}

func TestHandleCommandEcho(t *testing.T) {
	got := handleCommand(commandValue("ECHO", "hello"), NewStore())
	if got.typ != typeBulk || got.bulk != "hello" {
		t.Fatalf("ECHO = %#v, want hello bulk string", got)
	}
}

func TestHandleCommandSetGet(t *testing.T) {
	store := NewStore()

	got := handleCommand(commandValue("SET", "name", "Ahmed"), store)
	if got.typ != typeString || got.str != "OK" {
		t.Fatalf("SET = %#v, want OK", got)
	}

	got = handleCommand(commandValue("GET", "name"), store)
	if got.typ != typeBulk || got.bulk != "Ahmed" {
		t.Fatalf("GET existing = %#v, want Ahmed", got)
	}

	got = handleCommand(commandValue("GET", "missing"), store)
	if got.typ != typeNull {
		t.Fatalf("GET missing = %#v, want null", got)
	}
}

func TestHandleCommandMSetMGet(t *testing.T) {
	store := NewStore()

	got := handleCommand(commandValue("MSET", "name", "Ahmed", "city", "Hyderabad"), store)
	if got.typ != typeString || got.str != "OK" {
		t.Fatalf("MSET = %#v, want OK", got)
	}

	got = handleCommand(commandValue("MGET", "name", "missing", "city"), store)
	if got.typ != typeArray || len(got.array) != 3 {
		t.Fatalf("MGET = %#v, want three item array", got)
	}
	if got.array[0].typ != typeBulk || got.array[0].bulk != "Ahmed" {
		t.Fatalf("MGET first value = %#v, want Ahmed", got.array[0])
	}
	if got.array[1].typ != typeNull {
		t.Fatalf("MGET missing value = %#v, want null", got.array[1])
	}
	if got.array[2].typ != typeBulk || got.array[2].bulk != "Hyderabad" {
		t.Fatalf("MGET third value = %#v, want Hyderabad", got.array[2])
	}
}

func TestHandleCommandIncrDecr(t *testing.T) {
	store := NewStore()

	got := handleCommand(commandValue("INCR", "counter"), store)
	if got.typ != typeInteger || got.num != 1 {
		t.Fatalf("INCR missing = %#v, want 1", got)
	}

	got = handleCommand(commandValue("INCR", "counter"), store)
	if got.typ != typeInteger || got.num != 2 {
		t.Fatalf("INCR existing = %#v, want 2", got)
	}

	got = handleCommand(commandValue("DECR", "counter"), store)
	if got.typ != typeInteger || got.num != 1 {
		t.Fatalf("DECR existing = %#v, want 1", got)
	}

	store.Set("bad", "not-a-number")
	got = handleCommand(commandValue("INCR", "bad"), store)
	if got.typ != typeError {
		t.Fatalf("INCR non-integer = %#v, want error", got)
	}
}

func TestHandleCommandIncrKeepsTTL(t *testing.T) {
	now := time.Date(2026, time.April, 24, 12, 0, 0, 0, time.UTC)
	store := testStore(now)

	got := handleCommand(commandValue("SET", "counter", "1", "EX", "10"), store)
	if got.typ != typeString || got.str != "OK" {
		t.Fatalf("SET EX = %#v, want OK", got)
	}

	got = handleCommand(commandValue("INCR", "counter"), store)
	if got.typ != typeInteger || got.num != 2 {
		t.Fatalf("INCR = %#v, want 2", got)
	}

	got = handleCommand(commandValue("TTL", "counter"), store)
	if got.typ != typeInteger || got.num != 10 {
		t.Fatalf("TTL after INCR = %#v, want 10", got)
	}
}

func TestHandleCommandSetWithTTL(t *testing.T) {
	now := time.Date(2026, time.April, 24, 12, 0, 0, 0, time.UTC)
	store := testStore(now)

	got := handleCommand(commandValue("SET", "name", "Ahmed", "EX", "10"), store)
	if got.typ != typeString || got.str != "OK" {
		t.Fatalf("SET EX = %#v, want OK", got)
	}

	got = handleCommand(commandValue("TTL", "name"), store)
	if got.typ != typeInteger || got.num != 10 {
		t.Fatalf("TTL before expiry = %#v, want 10", got)
	}

	got = handleCommand(commandValue("GET", "name"), store)
	if got.typ != typeBulk || got.bulk != "Ahmed" {
		t.Fatalf("GET before expiry = %#v, want Ahmed", got)
	}

	store.now = func() time.Time { return now.Add(11 * time.Second) }

	got = handleCommand(commandValue("GET", "name"), store)
	if got.typ != typeNull {
		t.Fatalf("GET after expiry = %#v, want null", got)
	}

	got = handleCommand(commandValue("TTL", "name"), store)
	if got.typ != typeInteger || got.num != -2 {
		t.Fatalf("TTL after expiry = %#v, want -2", got)
	}
}

func TestHandleCommandExpireAndPersist(t *testing.T) {
	now := time.Date(2026, time.April, 24, 12, 0, 0, 0, time.UTC)
	store := testStore(now)
	store.Set("name", "Ahmed")

	got := handleCommand(commandValue("TTL", "name"), store)
	if got.typ != typeInteger || got.num != -1 {
		t.Fatalf("TTL without expiry = %#v, want -1", got)
	}

	got = handleCommand(commandValue("EXPIRE", "name", "20"), store)
	if got.typ != typeInteger || got.num != 1 {
		t.Fatalf("EXPIRE existing = %#v, want 1", got)
	}

	got = handleCommand(commandValue("TTL", "name"), store)
	if got.typ != typeInteger || got.num != 20 {
		t.Fatalf("TTL after EXPIRE = %#v, want 20", got)
	}

	got = handleCommand(commandValue("PERSIST", "name"), store)
	if got.typ != typeInteger || got.num != 1 {
		t.Fatalf("PERSIST existing = %#v, want 1", got)
	}

	got = handleCommand(commandValue("TTL", "name"), store)
	if got.typ != typeInteger || got.num != -1 {
		t.Fatalf("TTL after PERSIST = %#v, want -1", got)
	}

	got = handleCommand(commandValue("EXPIRE", "missing", "20"), store)
	if got.typ != typeInteger || got.num != 0 {
		t.Fatalf("EXPIRE missing = %#v, want 0", got)
	}
}

func TestHandleCommandRejectsInvalidTTL(t *testing.T) {
	tests := []struct {
		name  string
		value Value
	}{
		{name: "set negative", value: commandValue("SET", "name", "Ahmed", "EX", "-1")},
		{name: "set bad unit", value: commandValue("SET", "name", "Ahmed", "BAD", "10")},
		{name: "expire zero", value: commandValue("EXPIRE", "name", "0")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleCommand(tt.value, NewStore())
			if got.typ != typeError {
				t.Fatalf("handleCommand() = %#v, want error", got)
			}
		})
	}
}

func TestHandleCommandDeleteAndExists(t *testing.T) {
	store := NewStore()
	store.Set("name", "Ahmed")
	store.Set("city", "Hyderabad")

	got := handleCommand(commandValue("EXISTS", "name", "missing", "city"), store)
	if got.typ != typeInteger || got.num != 2 {
		t.Fatalf("EXISTS = %#v, want 2", got)
	}

	got = handleCommand(commandValue("DEL", "name", "missing"), store)
	if got.typ != typeInteger || got.num != 1 {
		t.Fatalf("DEL = %#v, want 1", got)
	}

	got = handleCommand(commandValue("EXISTS", "name", "city"), store)
	if got.typ != typeInteger || got.num != 1 {
		t.Fatalf("EXISTS after DEL = %#v, want 1", got)
	}
}

func TestHandleCommandErrors(t *testing.T) {
	tests := []struct {
		name  string
		value Value
	}{
		{name: "unknown command", value: commandValue("NOPE")},
		{name: "wrong arity", value: commandValue("GET")},
		{name: "not array", value: Value{typ: typeBulk, bulk: "PING"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleCommand(tt.value, NewStore())
			if got.typ != typeError || got.str == "" {
				t.Fatalf("handleCommand() = %#v, want error", got)
			}
		})
	}
}

func commandValue(parts ...string) Value {
	values := make([]Value, 0, len(parts))
	for _, part := range parts {
		values = append(values, Value{typ: typeBulk, bulk: part})
	}
	return Value{typ: typeArray, array: values}
}

func testStore(now time.Time) *Store {
	store := NewStore()
	store.now = func() time.Time { return now }
	return store
}

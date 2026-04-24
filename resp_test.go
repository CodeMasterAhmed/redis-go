package main

import (
	"strings"
	"testing"
)

func TestRespReadArray(t *testing.T) {
	resp := NewResp(strings.NewReader("*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n"))

	value, err := resp.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if value.typ != typeArray {
		t.Fatalf("typ = %q, want %q", value.typ, typeArray)
	}
	if len(value.array) != 2 {
		t.Fatalf("array length = %d, want 2", len(value.array))
	}
	if value.array[0].bulk != "ECHO" || value.array[1].bulk != "hello" {
		t.Fatalf("array = %#v, want ECHO hello", value.array)
	}
}

func TestRespReadBasicTypes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Value
	}{
		{name: "simple string", in: "+OK\r\n", want: Value{typ: typeString, str: "OK"}},
		{name: "error", in: "-ERR unknown command\r\n", want: Value{typ: typeError, str: "ERR unknown command"}},
		{name: "integer", in: ":12\r\n", want: Value{typ: typeInteger, num: 12}},
		{name: "bulk", in: "$5\r\nhello\r\n", want: Value{typ: typeBulk, bulk: "hello"}},
		{name: "null bulk", in: "$-1\r\n", want: Value{typ: typeNull}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewResp(strings.NewReader(tt.in)).Read()
			if err != nil {
				t.Fatalf("Read() error = %v", err)
			}
			if got.typ != tt.want.typ || got.str != tt.want.str || got.num != tt.want.num || got.bulk != tt.want.bulk {
				t.Fatalf("Read() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestRespReadRejectsInvalidBulkEnding(t *testing.T) {
	_, err := NewResp(strings.NewReader("$5\r\nhello\n")).Read()
	if err == nil {
		t.Fatal("Read() error = nil, want error")
	}
}

func TestValueMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "simple string", value: Value{typ: typeString, str: "OK"}, want: "+OK\r\n"},
		{name: "error", value: Value{typ: typeError, str: "ERR bad command"}, want: "-ERR bad command\r\n"},
		{name: "integer", value: Value{typ: typeInteger, num: 2}, want: ":2\r\n"},
		{name: "bulk", value: Value{typ: typeBulk, bulk: "hello"}, want: "$5\r\nhello\r\n"},
		{name: "null", value: Value{typ: typeNull}, want: "$-1\r\n"},
		{name: "array", value: Value{typ: typeArray, array: []Value{
			{typ: typeBulk, bulk: "GET"},
			{typ: typeBulk, bulk: "name"},
		}}, want: "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.value.Marshal()); got != tt.want {
				t.Fatalf("Marshal() = %q, want %q", got, tt.want)
			}
		})
	}
}

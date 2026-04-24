package main

import (
	"net"
	"testing"
	"time"
)

func TestHandleConnectionProcessesMultipleCommands(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		handleConnection(serverConn, NewStore())
		close(done)
	}()

	deadline := time.Now().Add(time.Second)
	if err := clientConn.SetDeadline(deadline); err != nil {
		t.Fatalf("SetDeadline() error = %v", err)
	}

	responses := NewResp(clientConn)

	writeRequest(t, clientConn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAhmed\r\n")
	got, err := responses.Read()
	if err != nil {
		t.Fatalf("read SET response error = %v", err)
	}
	if got.typ != typeString || got.str != "OK" {
		t.Fatalf("SET response = %#v, want OK", got)
	}

	writeRequest(t, clientConn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")
	got, err = responses.Read()
	if err != nil {
		t.Fatalf("read GET response error = %v", err)
	}
	if got.typ != typeBulk || got.bulk != "Ahmed" {
		t.Fatalf("GET response = %#v, want Ahmed", got)
	}

	if err := clientConn.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handleConnection did not exit after client close")
	}
}

func writeRequest(t *testing.T, conn net.Conn, request string) {
	t.Helper()

	if _, err := conn.Write([]byte(request)); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
}

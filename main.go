package main

import (
	"io"
	"log"
	"net"
)

func main() {
	if err := listenAndServe(":6379", NewStore()); err != nil {
		log.Fatal(err)
	}
}

func listenAndServe(addr string, store *Store) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("Listening on %s", addr)
	return serve(listener, store)
}

func serve(listener net.Listener, store *Store) error {
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go handleConnection(conn, store)
	}
}

func handleConnection(conn net.Conn, store *Store) {
	defer conn.Close()

	resp := NewResp(conn)
	writer := NewWriter(conn)

	for {
		value, err := resp.Read()
		if err != nil {
			if err != io.EOF {
				log.Printf("read error from %s: %v", conn.RemoteAddr(), err)
			}
			return
		}

		if err := writer.Write(handleCommand(value, store)); err != nil {
			log.Printf("write error to %s: %v", conn.RemoteAddr(), err)
			return
		}
	}
}

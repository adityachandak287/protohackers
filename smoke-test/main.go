package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	LISTENER_TYPE = "tcp"
	DEFAULT_HOST  = "0.0.0.0"
	DEFAULT_PORT  = 8888
)

func main() {
	host := flag.String("host", DEFAULT_HOST, "Host to listen on")
	port := flag.Int("port", DEFAULT_PORT, "Port to listen on")

	flag.Parse()

	address := fmt.Sprintf("%s:%d", *host, *port)

	listen, err := net.Listen(LISTENER_TYPE, address)
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()
	log.Printf("Server listening at %s", address)

	connId := 1

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleRequest(conn, connId)
		connId += 1
	}
}

func handleRequest(conn net.Conn, id int) {
	log.Printf("[%d] Serving %s", id, conn.RemoteAddr().String())

	buffer := make([]byte, 1024)
	for {
		nBytes, err := conn.Read(buffer)

		if err != nil {
			if err != io.EOF {
				log.Printf("[%d] Error reading from connection: %s", id, err)
			} else {
				log.Printf("[%d] Encountered EOF: %s", id, err)
			}
			break
		}

		input := buffer[:nBytes]
		log.Printf("[%d] Read %d bytes: [%v]", id, len(input), input)
		log.Printf("[%d] Sending %d bytes: [%v]", id, len(input), input)
		conn.Write(input)
	}

	log.Printf("[%d] Closing connection", id)
	if err := conn.Close(); err != nil {
		log.Printf("[%d] Error writing to connection: %s", id, err)
	}
}

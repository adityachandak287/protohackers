package main

import (
	"bytes"
	"net"
	"testing"
)

func TestSmokeTest(t *testing.T) {
	server, client := net.Pipe()

	go handleRequest(server, 1)

	testInput := []byte("Hello, World!")

	client.Write(testInput)
	buffer := make([]byte, 1024)
	nBytes, err := client.Read(buffer)
	if err != nil {
		t.Fatal(err)
	}

	resBytes := buffer[:nBytes]
	if !bytes.Equal(resBytes, testInput) {
		t.Fatalf("Invalid response: response %v != expected %v", resBytes, testInput)
	}

	client.Close()
}

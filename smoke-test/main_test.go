package main

import (
	"bytes"
	"io"
	"net"
	"testing"
)

var testCases = []string{"", " ", "Hello, World!", "!12345"}

func TestSmokeTest(t *testing.T) {
	for _, tc := range testCases {
		t.Run("TestCase", func(t *testing.T) {
			runSmokeTest(t, []byte(tc))
		})
	}
}

func runSmokeTest(t *testing.T, testInput []byte) {
	server, client := net.Pipe()

	go handleRequest(server, 1)

	go client.Write(testInput)

	inputLen := len(testInput)
	buffer := make([]byte, inputLen)
	nBytes, err := io.ReadFull(client, buffer)
	if err != nil {
		t.Fatal(err)
	}

	resBytes := buffer[:nBytes]
	if !bytes.Equal(resBytes, testInput) {
		t.Fatalf("Invalid response: response %v != expected %v", resBytes, testInput)
	}

	client.Close()
}

func FuzzSmokeTest(f *testing.F) {
	// seed corpus
	for _, tc := range testCases {
		f.Add([]byte(tc))
	}

	f.Fuzz(func(t *testing.T, randomString []byte) {
		runSmokeTest(t, randomString)
	})
}

package main

import (
	"bytes"
	"net"
	"sync"
	"testing"
)

type TestData struct {
	input  []byte
	output []byte
}

var testData = []TestData{
	{
		input:  []byte("{\"method\":\"isPrime\",\"number\":123}\n"),
		output: []byte("{\"method\":\"isPrime\",\"prime\":false}\n"),
	},
	{
		input:  []byte("{\"method\":\"isPrime\",\"number\":17}\n"),
		output: []byte("{\"method\":\"isPrime\",\"prime\":true}\n"),
	},
	{
		input:  []byte("{\"method\":\"isPrime\",\"number\":16.5}\n"),
		output: []byte("{\"method\":\"isPrime\",\"prime\":false}\n"),
	},
	{
		input:  []byte("{\"method\":\"isPrime\",\"number\":0}\n"),
		output: []byte("{\"method\":\"isPrime\",\"prime\":false}\n"),
	},
	{
		input:  []byte("{\"method\":\"isPrime\",\"number\":-7}\n"),
		output: []byte("{\"method\":\"isPrime\",\"prime\":false}\n"),
	},
	{
		input:  []byte("{\"method\":\"isPrime\"}\n"),
		output: []byte("{\"method\":\"isPrime\"}\n"),
	},
	{
		input:  []byte("{\"method\":\"isPrime\",\"number\":9322610.1234}\n"),
		output: []byte("{\"method\":\"isPrime\",\"prime\":false}\n"),
	},
}

func TestPrimeTimeBatch(t *testing.T) {
	for idx, data := range testData {
		var wg sync.WaitGroup
		server, client := net.Pipe()

		wg.Add(1)
		go func(conn net.Conn, connId int) {
			defer wg.Done()
			handleRequest(server, connId)
		}(server, idx+1)

		client.Write(data.input)
		buffer := make([]byte, 1024)
		nBytes, err := client.Read(buffer)
		if err != nil {
			t.Fatal(err)
		}

		resBytes := buffer[:nBytes]
		if !bytes.Equal(resBytes, data.output) {
			t.Fatalf("Invalid response: response %s != expected %s", resBytes, data.output)
		}

		client.Close()
		wg.Wait()
	}
}

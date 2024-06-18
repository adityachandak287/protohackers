package main

import (
	"bytes"
	"net"
	"reflect"
	"testing"
)

type Int32BytesTestData struct {
	Bytes []byte
	Int32 int32
}

func getInt32BytesTestData() []Int32BytesTestData {
	return []Int32BytesTestData{
		{
			// 0x3039 or 00000000000000000011000000111001
			Bytes: []byte{0, 0, 48, 57},
			Int32: 12345,
		},
		{
			Bytes: []byte{0, 0, 0, 0},
			Int32: 0,
		},
		{
			// FFFFCFC7 or 11111111111111111100111111000111
			Bytes: []byte{255, 255, 207, 199},
			Int32: -12345,
		},
		{
			// FFFFFFFF or 11111111111111111111111111111111
			Bytes: []byte{255, 255, 255, 255},
			Int32: -1,
		},
	}
}

func TestInt32FromBytes(t *testing.T) {
	for _, data := range getInt32BytesTestData() {
		if outInt32 := Int32FromBytes([4]byte(data.Bytes)); outInt32 != data.Int32 {
			t.Fatalf("Invalid int32 from bytes [Bytes %v] [OutputInt %d] != [ExpectedInt %d]", data.Bytes, outInt32, data.Int32)
		}
	}
}

func TestBytesFromInt32(t *testing.T) {
	for _, data := range getInt32BytesTestData() {
		if outBytes := BytesFromInt32(data.Int32); !bytes.Equal(outBytes[:], data.Bytes) {
			t.Fatalf("Invalid bytes from int32 [OutputBytes %v] [OutputInt %d] != [ExpectedBytes %v] [ExpectedInt %d]", outBytes, Int32FromBytes(outBytes), data.Bytes, data.Int32)
		}

	}
}

func formatMessage(msgType byte, timestamp int32, price int32) []byte {
	tsBytes := BytesFromInt32(timestamp)
	priceBytes := BytesFromInt32(price)

	buf := []byte{msgType}
	buf = append(buf, tsBytes[:]...)
	buf = append(buf, priceBytes[:]...)

	return buf
}

func TestFormatMessage(t *testing.T) {
	data := [][][]byte{
		{
			[]byte{73, 0, 0, 48, 57, 0, 0, 0, 101},
			formatMessage(InsertType, 12345, 101),
		},
		{
			[]byte{81, 0, 0, 3, 232, 0, 1, 134, 160},
			formatMessage(QueryType, 1000, 100000),
		},
	}

	for _, d := range data {
		expected := d[0]
		output := d[1]

		if !bytes.Equal(expected, output) {
			t.Fatalf("Expected %v != Output %v", expected, output)
		}
	}
}

func TestReadMessages(t *testing.T) {
	server, client := net.Pipe()

	messagesChan := make(chan interface{}, 10)

	go readMessages(server, 1, messagesChan)

	data := []struct {
		messageBytes []byte
		message      interface{}
	}{
		{
			messageBytes: formatMessage(InsertType, 12345, 101),
			message:      InsertMessage{InsertType, 12345, 101},
		},
		{
			messageBytes: formatMessage(InsertType, 12346, 102),
			message:      InsertMessage{InsertType, 12346, 102},
		},
		{
			messageBytes: formatMessage(InsertType, 12347, 100),
			message:      InsertMessage{InsertType, 12347, 100},
		},
		{
			messageBytes: formatMessage(InsertType, 40960, 5),
			message:      InsertMessage{InsertType, 40960, 5},
		},
		{
			messageBytes: formatMessage(QueryType, 12288, 16384),
			message:      QueryMessage{QueryType, 12288, 16384},
		},
	}

	for _, d := range data {
		client.Write(d.messageBytes)
		msg := <-messagesChan

		if !reflect.DeepEqual(msg, d.message) {
			t.Fatalf("Message mismatch! Expected [%T %v] != Output [%T %v]", msg, msg, d.message, d.message)
		}
	}

	client.Close()
}

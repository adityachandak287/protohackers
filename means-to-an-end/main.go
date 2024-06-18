package main

import (
	"encoding/binary"
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

	var tsdb TSDB = NewMapTSDB()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleRequest(conn, connId, tsdb)
		connId += 1
	}
}

type InsertMessage struct {
	Type      uint8
	Timestamp int32
	Price     int32
}

type QueryMessage struct {
	Type    uint8
	MinTime int32
	MaxTime int32
}

const (
	InsertType = byte('I')
	QueryType  = byte('Q')
	MessageLen = 9
	Int32Len   = 4
)

type Int32Bytes = [Int32Len]byte

func handleRequest(conn net.Conn, id int, db TSDB) {
	log.Printf("[%d] Serving %s", id, conn.RemoteAddr().String())

	dbId := fmt.Sprintf("%d", id)
	db.RegisterId(dbId)

	messagesChan := make(chan interface{}, 10)
	go readMessages(conn, id, messagesChan)

	for m := range messagesChan {
		switch msg := m.(type) {
		case InsertMessage:
			db.Insert(dbId, msg.Timestamp, msg.Price)
		case QueryMessage:
			avg, err := db.QueryAvg(dbId, msg.MinTime, msg.MaxTime)
			if err != nil {
				log.Printf("[%d] Error while querying average from tsdb: %s", id, err)
			}
			avgBytes := BytesFromInt32(avg)
			conn.Write(avgBytes[:])
		}
	}

	log.Printf("[%d] Closing connection", id)
	if err := conn.Close(); err != nil {
		log.Printf("[%d] Error writing to connection: %s", id, err)
	}
}

func readMessages(conn net.Conn, id int, ch chan interface{}) {
	defer close(ch)
	buffer := make([]byte, MessageLen)

	for loop := true; loop; {
		nBytes, err := io.ReadFull(conn, buffer)

		if err != nil {
			if err != io.EOF {
				log.Printf("[%d] Error reading from connection: %s", id, err)
			} else {
				log.Printf("[%d] Encountered EOF: %s", id, err)
			}
			break
		}

		if nBytes != len(buffer) {
			log.Fatalf("Could not read full buffer! %v", buffer)
		}

		log.Printf("[%d] Read %d bytes: [%v]", id, len(buffer), buffer)

		var msg interface{}
		switch messageType := buffer[0]; messageType {
		case InsertType:
			msg = InsertMessage{
				Type:      InsertType,
				Timestamp: Int32FromBytes(Int32Bytes(buffer[1 : 1+Int32Len])),
				Price:     Int32FromBytes(Int32Bytes(buffer[1+Int32Len : MessageLen])),
			}
		case QueryType:
			msg = QueryMessage{
				Type:    QueryType,
				MinTime: Int32FromBytes(Int32Bytes(buffer[1 : 1+Int32Len])),
				MaxTime: Int32FromBytes(Int32Bytes(buffer[1+Int32Len : MessageLen])),
			}
		default:
			log.Printf("[%d] Invalid message type: [%c]", id, messageType)
			loop = false
			continue
		}

		log.Printf("[%d] Parsed message: [%T %v]", id, msg, msg)
		ch <- msg
	}
}

func Int32FromBytes(buf Int32Bytes) int32 {
	binaryUint32 := binary.BigEndian.Uint32(buf[:])
	return int32(binaryUint32)

	/*
		// Non negative number
		if bits.LeadingZeros32(uint32(buf[0])) > 0 {
			return int32(binaryUint32)
		}

		// Negative number
		minusOne := ^uint32(0)                // https://groups.google.com/g/golang-nuts/c/2L7NfqtZYls/m/X8CLwdSEcu0J
		complement := binaryUint32 ^ minusOne // Flip all bits
		return (-1 * int32(complement+1))     // Add 1 and then negate
	*/
}

func BytesFromInt32(value int32) Int32Bytes {
	buf := make([]byte, Int32Len)
	binary.BigEndian.PutUint32(buf, uint32(value))
	return Int32Bytes(buf)
}

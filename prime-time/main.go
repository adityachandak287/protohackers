package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
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

type PrimeTimeRequest struct {
	Method string   `json:"method"`
	Number *float64 `json:"number,omitempty"`
}

type PrimeTimeResponse struct {
	Method string `json:"method"`
	Prime  *bool  `json:"prime,omitempty"`
}

func handleRequest(conn net.Conn, id int) {
	log.Printf("[%d] Serving %s", id, conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)

	for {
		input, err := reader.ReadBytes(byte('\n'))

		if err != nil {
			if err != io.EOF {
				log.Printf("[%d] Error reading from connection: %s", id, err)
				sendMalformedResponse(conn, id)
			} else {
				log.Printf("[%d] Encountered EOF: %s", id, err)
			}
			break
		}

		log.Printf("[%d] Input: %s", id, input)

		req := PrimeTimeRequest{}
		err = json.Unmarshal(input, &req)
		if err != nil {
			log.Printf("[%d] Error unmarshalling request data [%s]: %s", id, input, err)
			sendMalformedResponse(conn, id)
			break
		}

		log.Printf("[%d] Request data: %+v", id, req)

		resData := PrimeTimeResponse{Method: req.Method, Prime: nil}
		if req.Number != nil {
			isNumberPrime := new(bool)
			// Check if number is an integer
			if math.Floor(*req.Number) == *req.Number {
				numberInt := int64(math.Floor(*req.Number))
				*isNumberPrime = big.NewInt(numberInt).ProbablyPrime(0)
			} else {
				*isNumberPrime = false
			}
			resData.Prime = isNumberPrime
			log.Printf("[%d] Is %d Prime?: %v", id, req.Number, *isNumberPrime)
		} else {
			sendMalformedResponse(conn, id)
			break
		}
		log.Printf("[%d] Response data: %+v", id, resData)

		res, err := json.Marshal(resData)
		if err != nil {
			log.Printf("[%d] Error marshalling response data: %s", id, err)
			sendMalformedResponse(conn, id)
			break
		}
		res = append(res, '\n')
		conn.Write(res)
	}

	log.Printf("[%d] Closing connection", id)
	if err := conn.Close(); err != nil {
		log.Printf("[%d] Error writing to connection: %s", id, err)
	}
}

func sendMalformedResponse(conn net.Conn, id int) {
	res := PrimeTimeResponse{Method: "isPrime", Prime: nil}
	resBytes, _ := json.Marshal(res)
	resBytes = append(resBytes, '\n')
	log.Printf("[%d] Sending malformed response: %s", id, resBytes)
	conn.Write(resBytes)
}

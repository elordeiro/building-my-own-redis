package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	fmt.Println("Listening on port :6379")

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	defer conn.Close()

	for {
		resp := NewResp(conn)
		value, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		_ = value
		// fmt.Println(value)
		writer := NewWriter(conn)
		writer.Write(Value{typ: "string", str: "OK"})
		// conn.Write([]byte("+PONG\r\n"))
	}

}

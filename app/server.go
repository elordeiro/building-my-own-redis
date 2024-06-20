package main

import (
	"fmt"
	"net"
	"os"
)

type Server struct {
	Listener net.Listener
	Conn     []net.Conn
}

func NewServer() (*Server, error) {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		return nil, err
	}
	return &Server{
		Listener: l,
		Conn:     make([]net.Conn, 0),
	}, nil
}

func (s *Server) serverAccept() {
	conn, err := s.Listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		return
	}
	s.Conn = append(s.Conn, conn)

	go s.handleConnection(conn)
}

func (s *Server) serverClose() {
	for _, conn := range s.Conn {
		conn.Close()
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	for {
		resp := NewBuffer(conn)

		parsedResp, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			fmt.Println("Closing")
			return
		}

		result := Handler(parsedResp)
		writer := NewWriter(conn)
		writer.Write(result)
	}
}

func main() {
	server, err := NewServer()
	if err != nil {
		fmt.Println("Failed to create server")
		os.Exit(1)
	}

	fmt.Println("listening on port 6379")

	defer server.serverClose()

	for {
		server.serverAccept()
	}
}
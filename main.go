package main

import (
	"fmt"
	"io"
	"net"
)

const BufferSize = 2048

type Message struct {
	From string
	Data []byte
}

type Server struct {
	Host   string
	Port   int
	Ln     net.Listener
	Msgch  chan Message
	quitch chan struct{}
}

func NewServer(host string, port int) *Server {
	return &Server{
		Host:   host,
		Port:   port,
		Msgch:  make(chan Message),
		quitch: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", s.Port))
	if err != nil {
		fmt.Println("Start Listen Err: ", err.Error())
		return err
	}
	defer ln.Close()
	s.Ln = ln

	fmt.Println("Start Listen")
	go s.AcceptListen()

	<-s.quitch
	close(s.Msgch)

	return nil
}

func (s *Server) AcceptListen() {
	for {
		conn, err := s.Ln.Accept()
		if err != nil {
			fmt.Println("Accept Listen Err: ", err.Error())
			continue
		}

		fmt.Println("New connection to the server: ", conn.RemoteAddr().String())

		go s.Read(conn) // Can lead to race condition if there is computation logic
	}
}

func (s *Server) Read(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, BufferSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected:", conn.RemoteAddr())
				return
			}
			fmt.Println("Read error:", err)
			continue
		}

		msg := buf[:n]
		s.Msgch <- Message{
			From: conn.RemoteAddr().String(),
			Data: msg,
		}

		conn.Write([]byte("Thank you for your message!\n"))
	}
}

func main() {
	server := NewServer("localhost", 3000)

	go func() {
		for msg := range server.Msgch {
			fmt.Println(msg.From + ": " + string(msg.Data))
		}
	}()

	err := server.Start()
	if err != nil {
		fmt.Println(err.Error())
	}
}

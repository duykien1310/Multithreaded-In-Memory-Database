package main

import (
	"backend/internal/proto/echo"
	"backend/internal/server"
	"flag"
	"log"
	"net"
	"strconv"
)

func main() {
	host := flag.String("host", "0.0.0.0", "listen host")
	port := flag.Int("port", 6380, "listen port")
	flag.Parse()

	addr := net.JoinHostPort(*host, strconv.Itoa(*port))

	s, err := server.New(addr, echo.NewHandler())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("I/O multiplexing server listening on %s", addr)
	if err := s.Serve(); err != nil {
		log.Fatal(err)
	}
}

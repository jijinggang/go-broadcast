package main

import (
	"io"
	"log"
	"net"
)

type Server struct {
	addr    string
	run     bool
	clients map[net.Conn]bool
}

func NewServer(addr string) *Server {
	server := new(Server)
	server.addr = addr
	server.clients = make(map[net.Conn]bool)
	return server
}

func (server *Server) Start() {
	ln, err := net.Listen("tcp", server.addr)
	if err != nil {
		return
	}
	log.Printf("Server started on %s\n", server.addr)
	server.run = true
	for server.run {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go server.HandleConnection(conn)
	}
	ln.Close()
}

func (server *Server) BroadcastPacket(conn net.Conn, packet []byte) {
	for c := range server.clients {
		if c == conn {
			continue
		}
		go c.Write(packet)
	}
}

func (server *Server) HandleConnection(conn net.Conn) {
	addr := conn.RemoteAddr()
	log.Printf("Client connected from %s\n", addr.String())
	server.clients[conn] = true
	for {
		packet, err := ReadPacket(conn)
        log.Print(string(packet[4:]))
		if err != nil {
			break
		}
		go server.BroadcastPacket(conn, packet)
	}
	log.Printf("Client disconnected from %s\n", addr.String())
	delete(server.clients, conn)
	conn.Close()
}

func ReadPacket(r io.Reader) (packet []byte, err error) {
	size_buf := make([]byte, 4)
	n, err := io.ReadFull(r, size_buf)
	if err != nil {
		log.Printf("%d bytes read.\n", n)
		return nil, err
	}
	l := 0
	for i := 0; i < 4; i++ {
		l = (l << 8) + int(size_buf[i])
	}
	data_buf := make([]byte, l)
	n, err = io.ReadFull(r, data_buf)
	if err != nil {
		return nil, err
	}
	packet = make([]byte, 4+n)
	copy(packet, size_buf)
	copy(packet[4:], data_buf)
	return packet, nil
}

func main() {
	s := NewServer(":52234")
	s.Start()
}

package main

import (
        "io"
        "log"
        "net"
		"errors"
)

type Server struct {
        addr    string
        run     bool
        clients map[*Client]bool
}

type Client struct {
        conn    net.Conn
        packets chan []byte
        server  *Server
}

func NewClient(conn net.Conn, server *Server) *Client {
        client := new(Client)
        client.conn = conn
        client.server = server
        client.packets = make(chan []byte)
        return client
}

func NewServer(addr string) *Server {
        server := new(Server)
        server.addr = addr
        server.clients = make(map[*Client]bool)
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

func (server *Server) BroadcastPacket(client *Client, packet []byte) {
        for c := range server.clients {
                /*
                   if c == client {
                       continue
                   }
                */
                go c.SendPacket(packet)
        }
}

func (client *Client) SendPacket(packet []byte) {
        client.packets <- packet
}

func (client *Client) StartReceive() {
        for {
                packet, err := ReadPacket(client.conn)
                if err != nil {
                        break
                }
                go client.server.BroadcastPacket(client, packet)
        }
        log.Printf("Client disconnected from %s\n", client.conn.RemoteAddr().String())
        delete(client.server.clients, client)
        client.conn.Close()
}

func (client *Client) StartSend() {
        for {
                packet := <-client.packets
                _, err := client.conn.Write(packet)
                if err != nil {
                        break
                }
                log.Printf("Message dilivered to %s\n", client.conn.RemoteAddr().String())
        }
}

func (server *Server) HandleConnection(conn net.Conn) {
        addr := conn.RemoteAddr()
        client := NewClient(conn, server)
        log.Printf("Client connected from %s\n", addr.String())
        server.clients[client] = true
        go client.StartReceive()
        go client.StartSend()
}

//DataLen (4Byte)  MsgType(4Byte)  Data
func ReadPacket(r io.Reader) (packet []byte, err error) {

        size_buf := make([]byte, 4 + 4)
        n, err := io.ReadFull(r, size_buf)
        if err != nil || n < 4 + 4 {
                log.Printf("%d bytes read.\n", n)
                return nil, err
        }

        l := 0
        for i := 0; i < 4; i++ {
                l = (l << 8) + int(size_buf[i])
        }
		if l > 65535 {
			return nil, errors.New("too long buffer len")
		}
		t := 0
		for i:= 4; i< 8; i++{
			t =  (t << 8) + int(size_buf[i])
		}
        data_buf := make([]byte, l)
        n, err = io.ReadFull(r, data_buf)
        if err != nil || n < l {
                return nil, err
        }
        packet = make([]byte, 4+4+n)
        copy(packet, size_buf)
        copy(packet[4+4:], data_buf)
	    log.Println("recv msg,type:",t/*," data:", string(packet[8:])*/)
        return packet, nil
}

func main() {
        s := NewServer(":8000")
        s.Start()
}

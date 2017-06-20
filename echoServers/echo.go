package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	server, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening on", server.Addr())

	for {
		go func(conn net.Conn, err error) {
			for {
				if err == nil {
					in := bufio.NewReader(conn)
					message, err := in.ReadString('\n')
					if err == nil {
						_, err = fmt.Fprint(conn, message)
						if err != nil {
							conn.Close()
							return
						}
					} else {
						conn.Close()
					}
				} else {
					conn.Close()
					return
				}
			}
		}(server.Accept())
	}
}

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

	for {
		go func(conn net.Conn, err error) {
			if err == nil {
				in := bufio.NewReader(conn)
				message, newErr := in.ReadString('\n')
				if newErr == nil {
					_, err = fmt.Fprint(conn, message)
					if err == nil {
						conn.Close()
						return
					}
				}
			}
			conn.Close()
			fmt.Println(err.Error())
		}(server.Accept())
	}
}

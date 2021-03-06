package networking

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "3569"
	CONN_TYPE = "tcp"
)

var ipTable = map[string]chan string{}

func setServer() {
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		log.Println("connection error:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()
	log.Println("listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("error accepting: ", err.Error())
			os.Exit(1)
		}

		c := make(chan string)
		ipTable[conn.RemoteAddr().String()] = c

		// Handle connections in a new goroutine.
		go handleRequest(conn, c)
	}
}

func handleRequest(conn net.Conn, chan_server chan string) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("error reading: %s", err.Error())
	}
	request := string(buf[:n])

	switch request {
	case "usage":
		out <- Event{
			Type: UsageRequested,
			Data: conn.RemoteAddr().String(),
		}
	default:
		s := strings.SplitN(request, "::", 2)
		cmd := s[0]
		args := s[1]
		switch cmd {
		case "eval":
			ss := strings.SplitN(args, "|:|flow-code|:|", 3)
			if len(ss) != 3 {
				out <- Event{
					Type: Error,
					Data: "bad format",
				}
				conn.Close()
			} else {
				out <- Event{
					Type: EvalRequested,
					Data: map[string]string{
						"peer": conn.RemoteAddr().String(),
						"code": ss[1],
					},
				}
			}
		default:
			conn.Write([]byte("fuck you\n\n\a\a\a"))
			conn.Close()
			return
		}
	}

	go writeToConnection(conn, chan_server)
}

func writeToConnection(conn net.Conn, in <-chan string) {
	for s := range in {
		conn.Write([]byte(s))
	}
	conn.Close()
}

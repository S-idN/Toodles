package peer

import (
	"bufio"
	"fmt"
	"net"
)

func ReturnHashIdPortion() (hashIdPortion string) {
	test := "Test"
	return test
}

func StartServer(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("Failed start listener: %w", err)
	}
	defer listener.Close()

	fmt.Println("Listening on Port:", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go HandleConnection(conn)
	}
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		fmt.Println("Message:")
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed:", err)
			return
		}
		fmt.Println("Received message: ", message)
	}
}

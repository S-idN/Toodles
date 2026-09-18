package peer

import (
	"fmt"
	"net"
	"p2pChat/models"
	"strconv"
)

func ReturnHashIdPortion() (hashIdPortion string) {
	test := "Test"
	return test
}

func StartServer(port int, newNode *models.Node) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Failed start listener: %w", err)
	}
	defer listener.Close()

	fmt.Println("Listening on Port:", port)

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		newNode.Conn = conn

		go ReadFromPeer(conn)
		go WriteToPeer(conn)

		break
	}
}

// func HandleConnection(conn net.Conn) {
// 	defer conn.Close()

// 	reader := bufio.NewReader(conn)
// 	for {
// 		fmt.Println("Message:")
// 		message, err := reader.ReadString('\n')
// 		if err != nil {
// 			fmt.Println("Connection closed:", err)
// 			return
// 		}
// 		fmt.Println("Received message: ", message)
// 	}
// }

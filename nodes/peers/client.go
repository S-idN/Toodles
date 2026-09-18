package peer

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func ConnectToPeer(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return nil, fmt.Errorf("Could not connect to peer: %w", err)
	}
	return conn, nil
}

func WriteToPeer(conn net.Conn) error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text() + "\n"
		_, err := conn.Write([]byte(message))

		if err != nil {
			return fmt.Errorf("Error sending message to Peer")
		}
	}
	return scanner.Err()
}

func ReadFromPeer(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("Peer:", scanner.Text())
	}
	fmt.Println("Peer disconnected.")
}

func SendMessage(conn net.Conn, message string) error {
	_, err := conn.Write([]byte(message + "\n"))
	return err
}

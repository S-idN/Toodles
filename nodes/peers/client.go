package peer

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"p2pChat/models"
)

func ConnectToPeer(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return nil, fmt.Errorf("Could not connect to peer: %w", err)
	}
	return conn, nil
}

func ReadFromPeer(addr string, conn net.Conn, newNode *models.Node) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Printf("[%s]: %s\n", addr, scanner.Text())
	}
	fmt.Println("Peer disconnected:", addr)
	newNode.RemovePeer(addr)
}

func WriteLoop(newNode *models.Node) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		newNode.Broadcast(message)
	}
}

func SendMessage(conn net.Conn, message string) error {
	_, err := conn.Write([]byte(message + "\n"))
	return err
}

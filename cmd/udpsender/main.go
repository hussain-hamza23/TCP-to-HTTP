package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)


const address string = "localhost:42069"

func main(){
	conn, err := net.ResolveUDPAddr("udp", address)
	if err != nil{
		fmt.Printf("Error resolving address: %v\n", err)
		os.Exit(1)
	}

	socket, err := net.DialUDP("udp", nil, conn)
	if err != nil{
		fmt.Printf("Error dialing UDP: %v\n", err)
		os.Exit(1)
	}
	defer socket.Close()

	fmt.Printf("Sending messages to %s...\n", address)

	text := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		line, err := text.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		_, err = socket.Write([]byte(line))
		if err != nil {
			fmt.Printf("Error sending message: %v\n", err)
		}
	}
}
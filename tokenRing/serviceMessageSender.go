package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

type serviceMessage struct {
	TypeMessage string
	Dst         int
	Data        string
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func sendMessage(message []byte, myAddress string, recieverAddress string) {
	ServerAddr, err := net.ResolveUDPAddr("udp", recieverAddress)
	CheckError(err)
	LocalAddr, err := net.ResolveUDPAddr("udp", myAddress)
	CheckError(err)
	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	x, err := Conn.Write(message)
	if err != nil {
		fmt.Println(x, err)
		fmt.Println(message, err)
	}

	Conn.Close()
}

func main() {
	// startPort
	// TypeMessage
	// Data
	// Dst
	// Whome

	if len(os.Args) != 6 {
		fmt.Println("Неверное кол-во параметров")
		os.Exit(0)
	}

	startPort, _ := strconv.Atoi(os.Args[1])
	dst, _ := strconv.Atoi(os.Args[4])
	whome, _ := strconv.Atoi(os.Args[5])

	sMessage := serviceMessage{os.Args[2], dst, os.Args[3]}
	sMessageJSON, _ := json.Marshal(sMessage)
	fmt.Println(sMessage)
	sendMessage(sMessageJSON, "127.0.0.1:5000", "127.0.0.1:"+strconv.Itoa(startPort+whome))
}

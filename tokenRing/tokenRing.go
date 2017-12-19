package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type serviceMessage struct {
	TypeMessage string
	Dst         int
	Data        string
}

type Message struct {
	Type   string
	Data   string
	Sender int
	Dst    int
}

type Token struct {
	Origin int
	Mess   Message
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

func serviceRoutine(ServiceServerConn *net.UDPConn, serviceChannel chan serviceMessage) {
	for {
		buf := make([]byte, 1024)
		size, _, _ := ServiceServerConn.ReadFromUDP(buf)
		if size != 0 {
			var rec serviceMessage
			json.Unmarshal(buf[:size], &rec)
			serviceChannel <- rec
		}
	}
}

func listenTokenRoutine(ListenServerConn *net.UDPConn, listenChannel chan Token) {
	for {
		buf := make([]byte, 1024)
		size, _, _ := ListenServerConn.ReadFromUDP(buf)
		if size != 0 {
			var rec Token
			json.Unmarshal(buf[:size], &rec)
			listenChannel <- rec
		}
	}
}

func loop(myIndex int, numberNodes int, timeout int) {
	ServiceServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+strconv.Itoa(4000+myIndex))
	CheckError(err)
	ServiceServerConn, err := net.ListenUDP("udp", ServiceServerAddr)
	CheckError(err)
	defer ServiceServerConn.Close()
	serviceChannel := make(chan serviceMessage)
	go serviceRoutine(ServiceServerConn, serviceChannel)

	ListenServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+strconv.Itoa(3000+myIndex))
	CheckError(err)
	ListenServerConn, err := net.ListenUDP("udp", ListenServerAddr)
	CheckError(err)
	defer ListenServerConn.Close()
	listenChannel := make(chan Token)
	go listenTokenRoutine(ListenServerConn, listenChannel)

	if myIndex == 0 {
		send := Token{0, Message{"empty", "", 0, 0}}
		tokenJSON, _ := json.Marshal(send)
		sendMessage(tokenJSON, "127.0.0.1:"+strconv.Itoa(5000+myIndex), "127.0.0.1:"+strconv.Itoa(3000+(myIndex+1)%numberNodes))
		fmt.Println("Node", myIndex, ": Token Started")
	}

	waitTime := time.Millisecond * time.Duration((timeout+10)*(2*numberNodes))
	timer := time.NewTimer(waitTime)
	var recServiceMessage serviceMessage = serviceMessage{"empty", 0, ""}

	for {
		select {
		case recToken := <-listenChannel:
			time.Sleep(time.Millisecond * time.Duration(timeout))

			if recToken.Mess.Type == "electionResult" {
				fmt.Println("Node", myIndex, ": recieved token", recToken)

				if recToken.Mess.Dst == myIndex {
					fmt.Println("Node", myIndex, ": Election Done")
					send := Token{myIndex, Message{"empty", "", 0, 0}}
					tokenJSON, _ := json.Marshal(send)
					sendMessage(tokenJSON, "127.0.0.1:"+strconv.Itoa(5000+myIndex), "127.0.0.1:"+strconv.Itoa(3000+(myIndex+1)%numberNodes))
					fmt.Println("Node", myIndex, ": Token Started")

				} else {
					tokenJSON, _ := json.Marshal(recToken)
					sendMessage(tokenJSON, "127.0.0.1:"+strconv.Itoa(5000+myIndex), "127.0.0.1:"+strconv.Itoa(3000+(myIndex+1)%numberNodes))
				}

				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(waitTime)
				continue
			}

			if recToken.Mess.Type == "election" {
				fmt.Println("Node", myIndex, ": recieved token", recToken)
				if recToken.Mess.Dst == myIndex {
					data := strings.Split(recToken.Mess.Data, " ")
					data[len(data)-1] = strings.Replace(data[len(data)-1], "\n", "", -1)
					sort.Strings(data)
					where, _ := strconv.Atoi(data[0])
					recToken.Mess = Message{"electionResult", "", myIndex, where}
				} else {
					recToken.Mess.Data = recToken.Mess.Data + " " + strconv.Itoa(myIndex)
				}

				tokenJSON, _ := json.Marshal(recToken)
				sendMessage(tokenJSON, "127.0.0.1:"+strconv.Itoa(5000+myIndex), "127.0.0.1:"+strconv.Itoa(3000+(myIndex+1)%numberNodes))
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(waitTime)
				continue
			}

			if recServiceMessage.TypeMessage == "empty" {
				select {
				case recServiceMessage = <-serviceChannel:
					fmt.Println("Node", myIndex, ": recieved service message:", recServiceMessage)
				default:
				}
			}

			if recServiceMessage.TypeMessage == "drop" {
				recServiceMessage = serviceMessage{"empty", 0, ""}
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(waitTime)
				continue
			}

			fmt.Println("Node", myIndex, ": recieved token", recToken)

			if recToken.Mess.Type != "empty" {
				if recToken.Mess.Dst == myIndex && recToken.Mess.Type == "send" {
					recToken.Mess = Message{"notification", "", myIndex, recToken.Mess.Sender}
				}

				if recToken.Mess.Dst == myIndex && recToken.Mess.Type == "notification" {
					recToken.Mess = Message{"empty", "", 0, 0}
				}
			} else if recServiceMessage.TypeMessage == "send" {
				recToken.Mess = Message{"send", recServiceMessage.Data, myIndex, recServiceMessage.Dst}
				recServiceMessage = serviceMessage{"empty", 0, ""}
			}

			tokenJSON, _ := json.Marshal(recToken)
			sendMessage(tokenJSON, "127.0.0.1:"+strconv.Itoa(5000+myIndex), "127.0.0.1:"+strconv.Itoa(3000+(myIndex+1)%numberNodes))

			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(waitTime)
			continue
		case <-timer.C:
			fmt.Println("Node", myIndex, ": Timer Fired ON!")
			fmt.Println("Node", myIndex, ": Election started")
			send := Token{myIndex, Message{"election", strconv.Itoa(myIndex), 0, myIndex}}
			tokenJSON, _ := json.Marshal(send)
			sendMessage(tokenJSON, "127.0.0.1:"+strconv.Itoa(5000+myIndex), "127.0.0.1:"+strconv.Itoa(3000+(myIndex+1)%numberNodes))
			timer.Reset(waitTime)
		default:
		}

	}
}

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Неверное кол-во параметров")
		os.Exit(0)
	}

	numberNodes := 10
	timeout := 1

	for i := 1; i < len(os.Args); i += 2 {
		if os.Args[i] == "--n" {
			numberNodes, _ = strconv.Atoi(os.Args[i+1])
		}

		if os.Args[i] == "--t" {
			timeout, _ = strconv.Atoi(os.Args[i+1])
		}
	}

	fmt.Println(numberNodes, timeout)

	for i := 1; i < numberNodes; i++ {
		go loop(i, numberNodes, timeout)
	}

	time.Sleep(time.Microsecond * time.Duration(3))
	loop(0, numberNodes, timeout)
}

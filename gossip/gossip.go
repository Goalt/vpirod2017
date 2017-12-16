package main

import (
	"./graph"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

type Message struct {
	Id          int
	TypeMessage string
	Sender      int
	Origin      int
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

func recieveGoRoutine(N int, myIndex int, channelBetweenRecieveAndSend chan []byte, closeChannel chan string, breakChannel chan int, ServerAddr *net.UDPAddr, ServerConn *net.UDPConn) {
	dictOfRecievedMessages := make(map[string]bool)
	flag := 0

	if myIndex == 0 {
		flag = 1
		mess := Message{0, "multicast", 0, 0}
		message, _ := json.Marshal(mess)
		channelBetweenRecieveAndSend <- message
	}

	for {
		select {
		case <-closeChannel:
		default:
			buf := make([]byte, 1024)
			size, addr, _ := ServerConn.ReadFromUDP(buf)

			if size != 0 {
				var rec Message
				json.Unmarshal(buf[:size], &rec)

				if (rec.TypeMessage == "multicast") && (flag == 0) {
					fmt.Println("Received ", string(buf[:size]), " from ", addr, "who ", ServerAddr, "\n")

					flag = 1
					channelBetweenRecieveAndSend <- buf[:size]

					rec = Message{0, "notification", myIndex, myIndex}
					message, _ := json.Marshal(rec)
					channelBetweenRecieveAndSend <- message

				} else if rec.TypeMessage == "notification" {
					key := strconv.Itoa(rec.Id) + " " + strconv.Itoa(rec.Origin)
					_, ok := dictOfRecievedMessages[key]

					if ok == false {
						fmt.Println("Received ", string(buf[:size]), " from ", addr, "who ", ServerAddr, "\n")
						dictOfRecievedMessages[key] = true
						channelBetweenRecieveAndSend <- buf[:size]
					}
				}
			}

			if (myIndex == 0) && (len(dictOfRecievedMessages) == N-1) {
				breakChannel <- 1
			}
		}
	}
}

func sendGoRoutine(connectionGraph graph.Graph, startPort int, myIndex int, timeout int, ttl int, N int, ticksChannel chan int, closeChannel chan string, channelBetweenRecieveAndSend chan []byte) {
	myPort := startPort + myIndex
	myAddress := "127.0.0.1:" + strconv.Itoa(myPort)
	numberOfTicks := 0

	for {
		select {
		case <-ticksChannel:
			ticksChannel <- numberOfTicks
		case <-closeChannel:
		case msg := <-channelBetweenRecieveAndSend:
			var sen Message
			json.Unmarshal(msg, &sen)

			var neighbours []int
			neighboursNodes, _ := connectionGraph.Neighbors(myIndex)
			for i := range neighboursNodes {
				nei, _ := strconv.Atoi(neighboursNodes[i].String())
				neighbours = append(neighbours, nei)
			}

			index := -1
			for i, mas := range neighbours {
				if mas == sen.Sender {
					index = i
					break
				}
			}
			if index != -1 {
				neighbours = append(neighbours[:index], neighbours[index+1:]...)
			}

			sen.Sender = myIndex
			message, _ := json.Marshal(sen)

			if len(neighbours) == 0 {
				continue
			}

			for j := 0; j < ttl; j++ {
				neig := rand.Intn(len(neighbours))
				recieverAddress := "127.0.0.1:" + strconv.Itoa(N+startPort+neighbours[neig])
				sendMessage(message, myAddress, recieverAddress)
				fmt.Println("Sended ", string(message), " to ", recieverAddress, "from ", myAddress, "\n")

				numberOfTicks += 1
				time.Sleep(time.Millisecond * time.Duration(timeout))
			}

		default:
			numberOfTicks += 1
			time.Sleep(time.Millisecond * time.Duration(timeout))
		}
	}
}

func loopReceiveSend(connectionGraph graph.Graph, startPort int, myIndex int, timeout int, ttl int, N int) {
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+strconv.Itoa(N+myIndex+startPort))
	CheckError(err)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	channelBetweenRecieveAndSend := make(chan []byte)
	breakChannel := make(chan int)
	closeChannel := make(chan string)
	ticksChannel := make(chan int)

	go recieveGoRoutine(N, myIndex, channelBetweenRecieveAndSend, closeChannel, breakChannel, ServerAddr, ServerConn)
	go sendGoRoutine(connectionGraph, startPort, myIndex, timeout, ttl, N, ticksChannel, closeChannel, channelBetweenRecieveAndSend)

	<-breakChannel
	ticksChannel <- 0
	numberOfTicks := <-ticksChannel
	fmt.Println("Number of ticks: ", numberOfTicks)
	close(closeChannel)
}

func main() {

	// N - кол-во узлов
	// startPort - начальный порт
	// timeout - время ожидания для следующей рассылки
	// minDegree
	// maxDegree
	// ttl

	if len(os.Args) != 7 {
		fmt.Println("Неверное кол-во параметров")
		os.Exit(0)
	}

	rand.Seed(time.Now().UnixNano())

	N, _ := strconv.Atoi(os.Args[1])
	startPort, _ := strconv.Atoi(os.Args[2])
	timeout, _ := strconv.Atoi(os.Args[3])
	minDegree, _ := strconv.Atoi(os.Args[4])
	maxDegree, _ := strconv.Atoi(os.Args[5])
	ttl, _ := strconv.Atoi(os.Args[6])

	connectionGraph := graph.Generate(N, minDegree, maxDegree, 0)

	for i := 1; i < N; i++ {
		go loopReceiveSend(connectionGraph, startPort, i, timeout, ttl, N)
	}
	loopReceiveSend(connectionGraph, startPort, 0, timeout, ttl, N)

	fmt.Println("Connection Graph: ", connectionGraph)
}

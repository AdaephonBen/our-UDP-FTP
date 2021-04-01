package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"syscall"
)

const PACKET_SIZE = 10240
const FILE_SIZE = 104857600

func main() {
	var wg sync.WaitGroup
	sendSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	recvSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		fmt.Println("Error while creating socket", err)
	}
	sendAddress := syscall.SockaddrInet4{
		Port: 6790,
		Addr: [4]byte{192, 168, 1, 5},
	}

	recvAddress := syscall.SockaddrInet4{
		Port: 6789,
		Addr: [4]byte{192, 168, 1, 5},
	}
	err = syscall.Bind(recvSocket, &recvAddress)

	noPacketsReceived := 0

	file := make([]byte, FILE_SIZE)
	recvBuffer := make([]byte, 10246)

	hasBeenReceived := make([]bool, FILE_SIZE/PACKET_SIZE)

	for true {
		_, _, _ = syscall.Recvfrom(recvSocket, recvBuffer, 0)
		s := string(recvBuffer[0:6])
		i, _ := strconv.Atoi(s)

		wg.Add(1)
		go SendAck(&wg, sendSocket, recvBuffer[0:6], sendAddress)
		if !hasBeenReceived[i] {
			for j := i * PACKET_SIZE; j < (i+1)*PACKET_SIZE; j++ {
				file[j] = recvBuffer[j-i*PACKET_SIZE+6]
			}
			noPacketsReceived += 1
			hasBeenReceived[i] = true
			fmt.Println(noPacketsReceived)
			if noPacketsReceived == FILE_SIZE/PACKET_SIZE {
				break
			}
		}
	}
	wg.Wait()
	_ = ioutil.WriteFile("file", file, 0644)
}

func SendAck(wg *sync.WaitGroup, socket int, data []byte, sendAddress syscall.SockaddrInet4) {
	defer wg.Done()
	_ = syscall.Sendto(socket, data, 0, &sendAddress)
}

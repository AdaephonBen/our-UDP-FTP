package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"syscall"
	// "time"
)

const PACKET_SIZE = 10240
const FILE_SIZE = 104857600

var hasBeenReceived [FILE_SIZE / PACKET_SIZE]bool

func main() {
	sendSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	recvSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		fmt.Println("Error while creating socket", err)
	}

	file, err := ioutil.ReadFile("CS3543_100MB")
	fmt.Println(len(file))
	if err != nil {
		fmt.Println("Error while reading file", err)
	}

	sendAddress := syscall.SockaddrInet4{
		Port: 6789,
		Addr: [4]byte{192, 168, 1, 5},
	}

	recvAddress := syscall.SockaddrInet4{
		Port: 6790,
		Addr: [4]byte{192, 168, 1, 5},
	}
	acks := make(chan int, FILE_SIZE/PACKET_SIZE)

	_ = syscall.Bind(recvSocket, &recvAddress)

	for i := range hasBeenReceived {
		hasBeenReceived[i] = false
	}

	go recvAcks(recvSocket, 0, acks)

	for true {
		flag := true
		ackedPackets := 0
		for i := range hasBeenReceived {
			if !hasBeenReceived[i] {
				flag = false
				packet_number := []byte(strconv.Itoa(i))
				for len(packet_number) < 6 {
					packet_number = append([]byte{byte('0')}, packet_number...)
				}
				dataToSend := append(packet_number, file[i*PACKET_SIZE:(i+1)*PACKET_SIZE]...)
				go sendPacket(sendSocket, sendAddress, dataToSend)
				// time.Sleep(3 * time.Microsecond)
			} else {
				ackedPackets += 1
			}
			select {
			case i := <-acks:
				hasBeenReceived[i] = true
			default:
			}
		}
		fmt.Println(ackedPackets)
		if flag {
			break
		}
	}

	fmt.Println("Sent file")
}

func sendPacket(socket int, recvAddress syscall.SockaddrInet4, data []byte) {
	_ = syscall.Sendto(socket, data, 0, &recvAddress)
}

func recvAcks(socket int, acksReceived int, acks chan int) {
	recvBuffer := make([]byte, 6)
	_, _, _ = syscall.Recvfrom(socket, recvBuffer, 0)
	s := string(recvBuffer)
	i, _ := strconv.Atoi(s)
	select {
	case acks <- i:
	default:
	}
	// fmt.Println(acksReceived)
	// hasBeenReceived[i] = true
	// if acksReceived == FILE_SIZE/PACKET_SIZE {
	// 	return
	// }
	go recvAcks(socket, acksReceived+1, acks)
}

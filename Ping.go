package main

// Inspired by https://gist.github.com/lmas

import (
	"fmt"
	"net"
	"os"
	time "time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var defddr = "0.0.0.0"

func Ping(address string, timeLengt int) (*net.IPAddr, time.Duration, error) {
	connection, err := icmp.ListenPacket("ip4:icmp", defddr) //Listen to all replies
	if err != nil {
		panic(err)
		return nil, 0, err
	}

	// Step 0 : ARPit
	addr, err := net.ResolveIPAddr("ip4", address)
	if err != nil {
		panic(err)
		return addr, 0, err
	}
	// Step 1 create the mesage. Thank you Imas for this useful piece of code.
	println("Step1")
	ICMPmessage := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0, Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: []byte("")},
	}
	//Step 2 bytefy the message.
	println("Step2")

	binary, err := ICMPmessage.Marshal(nil)

	// Step 3 punch it chewie!
	println("Step3")
	startTime := time.Now()
	bites, err := connection.WriteTo(binary, addr)
	if err != nil {
		panic(err)
		return addr, 0, nil
	} else if bites != len(binary) {
		print("error")
		return nil, 0, nil
	}

	// Step 4 wait for reply ( If there is any )

	println("Step4")
	replySpace := make([]byte, 1500) // sisze of packet
	err = connection.SetDeadline(time.Now().Add(time.Duration(timeLengt) * time.Second))
	// catch if there is an error
	if err != nil {
		panic(err)
		return addr, 0, err
	}

	bites, peer, err := connection.ReadFrom(replySpace)
	dur := time.Since(startTime)
	if err != nil {
		panic(err)
		return addr, 0, err
	}
	returnMessage, err := icmp.ParseMessage(1, replySpace[:bites])
	if err != nil {
		panic(err)
		return addr, 0, err
	}
	if returnMessage.Type == ipv4.ICMPTypeEchoReply {
		println("END")
		return addr, dur, nil
	} else {
		println("END with Error")
		return addr, -1, fmt.Errorf("ERROR from %v", peer)
	}
}

func main() {
	fmt.Println("Hello world")
	dst, dur, err := Ping("www.googlsdfsdfsdfsdfsfe.com", 10)

	fmt.Printf("Ping to %s , TTL: %s Error code: %s\n", dst, dur, err)
}

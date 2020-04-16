package main

// Inspired by https://gist.github.com/lmas

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	time "time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var defddr = "0.0.0.0"
var success = 0
var lost = 0

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
	ICMPmessage := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0, Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: []byte("")},
	}
	//Step 2 bytefy the message.

	binary, err := ICMPmessage.Marshal(nil)

	// Step 3 punch it chewie!
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
		success++
		return addr, dur, nil
	} else {
		println("END with Error")
		lost++
		return addr, 0, fmt.Errorf("ERROR from %v", peer)
	}
}

func main() {
	var count float64 = 0
	var totalTime float64 = 0
	var averagetime float64 = 0

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() { // This function awaits for an interrupt to occure
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	fmt.Println("To exit press CTRL + C")

	go func() { // Runs Infinite loop for ping
		for {
			dst, dur, err := Ping("8.8.8.8", 1)
			fmt.Printf("Ping to %s , TTL: %s Error code: %s\n", dst, dur, err)
			count = count + 1 // counts howmany packets have been send
			/*
				This is chunk of code calculates the average of the trip
				vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
			*/
			ms := float64(dur / time.Millisecond)
			totalTime = totalTime + ms
			averagetime = float64(totalTime / count)
		}
	}()
	<-done
	fmt.Printf("Total packets send: %d Lost: %d Recieved: %d Average TTL: %f\n", lost+success, lost, success, averagetime)

}

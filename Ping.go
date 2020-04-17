package main

// Inspired by https://gist.github.com/lmas

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	time "time"

	"golang.org/x/net/icmp"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var defddr = "0.0.0.0"
var success = 0
var lost = 0

func Ping4(address string, timeLengt int) (*net.IPAddr, time.Duration, error) {
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
func Ping6(address string, timeLengt int) (string, time.Duration, error) {

	connection, err := net.Dial("ip6:58", address) //Listen to all replies I feel this is the line that has the error, but I dont know how to fix it.
	if err != nil {
		panic(err)
		return "", 0, err
	}

	// Step 1 create the mesage. Thank you Imas for this useful piece of code.
	ICMPmessage := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
	//Step 2 bytefy the message.

	binary, err := ICMPmessage.Marshal(nil)

	// Step 3 punch it chewie!
	startTime := time.Now()
	bites, err := connection.Write(binary)
	if err != nil {
		panic(err)
		return address, 0, nil
	} else if bites != len(binary) {
		print("error")
		return "", 0, nil
	}

	// Step 4 wait for reply ( If there is any )

	replySpace := make([]byte, 1500) // sisze of packet
	err = connection.SetDeadline(time.Now().Add(time.Duration(timeLengt) * time.Second))
	// catch if there is an error
	if err != nil {
		panic(err)
		return address, 0, err
	}

	bitesa, err := connection.Read(replySpace)
	dur := time.Since(startTime)
	if err != nil {
		panic(err)
		return address, 0, err
	}
	returnMessage, err := icmp.ParseMessage(1, replySpace[:bitesa])
	if err != nil {
		panic(err)
		return address, 0, err
	}
	if returnMessage.Type == ipv6.ICMPTypeEchoReply {
		println("END")
		success++
		return address, dur, nil
	} else {
		println("END with Error")
		lost++
		return address, 0, fmt.Errorf("ERROR")
	}
}

func pingerMethod(protocol bool, TTL int, address string) {
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
	if protocol == false {
		go func() { // Runs Infinite loop for ping
			for {
				dst, dur, err := Ping4(address, TTL)
				fmt.Printf("Ping to %s , RTT: %s Error code: %s\n", dst, dur, err)
				count = count + 1 // counts how many packets have been send
				/*
					This is chunk of code calculates the average of the trip
					vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
				*/
				ms := float64(dur / time.Millisecond)
				totalTime = totalTime + ms
				averagetime = float64(totalTime / count)

			}
		}()
	} else {
		go func() { // Runs Infinite loop for ping
			for {
				fmt.Println("This function does not work, I am not sure why yet\n I believe is because i can not have a reliable IPV6 return address\n I would require some assistance please")
				dst, dur, err := Ping6(address, TTL)
				fmt.Printf("Ping to %s , RTT: %s Error code: %s\n", dst, dur, err)
				count = count + 1 // counts how many packets have been send
				/*
					This is chunk of code calculates the average of the trip
					vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
				*/
				ms := float64(dur / time.Millisecond)
				totalTime = totalTime + ms
				averagetime = float64(totalTime / count)

			}
		}()
	}
	<-done // Magic, still dont understand why I need this line.
	fmt.Printf("Total packets send: %d Lost: %d Recieved: %d Average RLL: %f\n", lost+success, lost, success, averagetime)

}

func main() {

	/***
	 *    ______ _____ _   _ _____   _____ _____ _
	 *    | ___ \_   _| \ | |  __ \ |_   _|_   _| |
	 *    | |_/ / | | |  \| | |  \/   | |   | | | |
	 *    |  __/  | | | . ` | | __    | |   | | | |
	 *    | |    _| |_| |\  | |_\ \  _| |_  | | |_|
	 *    \_|    \___/\_| \_/\____/  \___/  \_/ (_)
	 *
	 *
		 ***/
	/*
		By: Enrique Calderon
		AKA: PointRain
		Disclaimer: I have not a lot of experience with Golang, I have learned a lot and I really liked the challenge
					I would lke to say that the IPV6 function does not work and I tried for a long time to make it work
					If you know how to make it work please contact me.

					Thank you
	*/
	IP := flag.String("IP", "nil", "Ip address")
	protocol := flag.Bool("IPV6", false, "True if its a IPV6 protocol")
	timeToLive := flag.Int("TTL", 1, "Sets time to live. Default is 1")
	flag.Parse()
	if strings.Compare(*IP, "nil") == 0 {
		fmt.Println("Please put an IP")
	} else if *timeToLive <= 0 {
		fmt.Println("Please put a positive and greater that 0 integer for TTL")
	} else {
		pingerMethod(*protocol, *timeToLive, *IP)
	}

}

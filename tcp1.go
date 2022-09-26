package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func readArgs(a []string) (map[int]string, [2]int, int) {

	id, _ := strconv.Atoi(a[0])

	f, err := os.Open("config.txt")

	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	data := make([]string, 0)

	for scanner.Scan() {
		data = append(data, scanner.Text())
	}

	d := strings.Split(data[0], " ")

	var delays [2]int
	delays[0], _ = strconv.Atoi(d[0])
	delays[1], _ = strconv.Atoi(d[1])

	ports := make(map[int]string)
	for i := 1; i < len(data); i++ {
		s := strings.Split(data[i], " ")
		ports[i] = s[1] + ":" + s[2]
	}

	return ports, delays, id

}

func unicast_receive(message string, ch chan string) {
	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"
	ch <- message + "@@@" + myTime

}

func handleConnection(c net.Conn, ch chan string) {
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			break
		}

		unicast_receive(temp, ch)

	}
	c.Close()
}

func client(ports map[int]string, delays [2]int, id int, ch chan string) {

	go waitForMessages(ch)
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')

		input := strings.Split(text, " ")

		if strings.ToLower(input[0]) == "send" {

			n, _ := strconv.Atoi(input[1])
			id_s := strconv.Itoa(id)

			message := id_s + "@@@" + strings.Join(input[2:], " ")

			//fmt.Println(message)

			r := rand.Intn(delays[1]) + delays[0]

			go unicast_send(n, ports[n], message, r)

		} else if strings.TrimSpace(string(text)) == "STOP" { //is this ok
			fmt.Println("TCP client exiting...")
			return
		}

	}

}

func unicast_send(n int, destination string, message string, r int) {

	time.Sleep(time.Duration(r) * time.Millisecond)

	m := strings.TrimSuffix(message, "\n")
	msg := strings.Split(m, "@@@")
	c, err := net.Dial("tcp", destination)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Fprintf(c, message+"\n")

	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"

	fmt.Println("Sent " + msg[1] + " to process " + strconv.Itoa(n) + ", system time is " + myTime)
}

func waitForMessages(ch chan string) {

	for {
		out := strings.Split(<-ch, "@@@")
		fmt.Println("Received " + out[1] + " from process " + out[0] + ", system time is " + out[2])
	}

}

func main() {

	a := os.Args[1:] //Take input into a

	if len(a) == 0 {
		fmt.Println("Please provide process number")
		return
	} //Ensure a process id is provided for

	ports, delays, id := readArgs(a)
	//it gets here
	PORT := ":" + strings.Split(ports[id], ":")[1]
	//Code from https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/#create-a-concurrent-tcp-server

	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer l.Close()

	ch := make(chan string)
	go client(ports, delays, id, ch) //will start first client

	for {

		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c, ch)
	}
}

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

func unicast_send(destination, message string) {
	msg := strings.Split(message, "\n")
	c, err := net.Dial("tcp", destination)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Fprintf(c, message+"\n")

	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"

	fmt.Println("Sent " + msg[1] + " to process " + destination + ", system time is " + myTime)
}

func unicast_receive(message string, m chan string) {
	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"
	m <- message + myTime
}

func launch_server(PORT string, m chan string, delays [2]int) {

	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		r := rand.Intn(delays[1]) + delays[0]
		time.Sleep(time.Duration(r) * time.Millisecond)
		unicast_receive(netData, m)
	}

}

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

func main() {
	a := os.Args[1:] //Take input into a

	if len(a) == 0 {
		fmt.Println("Please provide process number")
		return
	} //Ensure a process id is provided for

	ports, delays, id := readArgs(a)

	m := make(chan string)

	PORT := ":" + strings.Split(ports[id], ":")[1]

	go launch_server(PORT, m, delays)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')

		input := strings.Split(text, " ")

		if strings.ToLower(input[0]) == "send" {

			n, _ := strconv.Atoi(input[1])
			id_s := strconv.Itoa(id + 1)

			message := id_s + "\n" + input[2]

			r := rand.Intn(delays[1]) + delays[0]
			time.Sleep(time.Duration(r) * time.Millisecond)

			unicast_send(ports[n], message)

		} else if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
		if len(m) != 0 {
			out := strings.Split(<-m, "\n")

			fmt.Println("Received " + out[1] + " to process " + out[0] + ", system time is " + out[2])
		}

	}

}

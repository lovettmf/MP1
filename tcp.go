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

func unicast_send(destination, message string, id string, random int) {
	c, err := net.Dial("tcp", destination)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Time to sleep!")
	time.Sleep(time.Duration(random) * time.Millisecond)
	fmt.Println("I'm Awake!")

	fmt.Fprintf(c, id+"|"+message+"\n")

	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"

	fmt.Println("Sent " + message + " to " + destination + ", system time is " + myTime)
	c.Close()
}

func unicast_receive(source net.Conn, message chan string, delays [2]int) {
	r := rand.Intn(delays[1]) + delays[0]
	fmt.Println("Time to sleep")
	time.Sleep(time.Duration(r) * time.Millisecond)
	fmt.Println("I'm Awake!")

	netData, err := bufio.NewReader(source).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"

	temp := strings.Split(netData, "|")

	message <- strings.TrimSuffix(temp[1], "\n")
	message <- temp[0]
	message <- myTime

}

func launch_server(PORT string, m chan string, delays [2]int) {

	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		//fmt.Println("Checkpoint 1")
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		unicast_receive(c, m, delays)

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

	m := make(chan string, 3)

	PORT := ":" + strings.Split(ports[id], ":")[1]

	go launch_server(PORT, m, delays)

	for {
		fmt.Println("Waiting for input")

		reader := bufio.NewReader(os.Stdin)

		text, _ := reader.ReadString('\n')

		input := strings.Split(text, " ")

		if strings.ToLower(input[0]) == "send" {

			n, _ := strconv.Atoi(input[1])
			id_s := strconv.Itoa(id)

			message := strings.TrimSuffix(strings.Join(input[2:], " "), "\n")

			r := rand.Intn(delays[1]) + delays[0]

			go unicast_send(ports[n], message, id_s, r)

		} else if strings.TrimSpace(input[0]) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}

		fmt.Println("Waiting to receive")
		if len(m) == 3 {
			msg := <-m
			src := <-m
			t := <-m

			fmt.Println("Received " + msg + " from Process " + src + ", system time is " + t)
		}

	}

}

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

	//Process ID
	id, _ := strconv.Atoi(a[0])

	//Open and read the config file
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

	//message delay range
	var delays [2]int
	delays[0], _ = strconv.Atoi(d[0])
	delays[1], _ = strconv.Atoi(d[1])

	//map of ID/port pairs
	ports := make(map[int]string)
	for i := 1; i < len(data); i++ {
		s := strings.Split(data[i], " ")
		ports[i] = s[1] + ":" + s[2]
	}

	return ports, delays, id

}

func unicast_receive(message string, ch chan string) {

	//Time message is receiving by server
	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"

	//Places message and time in channel for receiving client
	//"@@@" will be used as a delimiting agent
	ch <- message + "@@@" + myTime
}

func handleConnection(c net.Conn, ch chan string) {
	//Uses code from https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/#create-a-concurrent-tcp-server

	//Handles each incoming TCP connection
	for {
		//Get TCP data. Note that anything after the first "\n" character will be disregarded
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		//Extract message string
		temp := strings.TrimSpace(string(netData))

		if temp == "STOP" {
			break
		}

		//Calls unicast_receive to deliver the message to a client
		unicast_receive(temp, ch)

	}
	c.Close() //Ends connection
}

func client(addresses map[int]string, delays [2]int, id int, ch chan string, exit chan bool) {
	//Main client will continuously take input
	//Launches new routine for each message that needs sending

	go waitForMessages(ch) //A routine to monitor for messages from the server

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')

		input := strings.Split(text, " ")

		if strings.ToLower(input[0]) == "send" {

			n, _ := strconv.Atoi(input[1]) //Process ID of destination
			id_s := strconv.Itoa(id)       //Own process ID

			//The senders process ID is added for future splitting by the recipient
			//Note use of delimiter @@@ because "\n" cannot be used
			message := id_s + "@@@" + strings.Join(input[2:], " ")

			//Random delay time
			r := rand.Intn(delays[1]) + delays[0]

			//Send is handled in its own routine
			go unicast_send(n, addresses[n], message, r)

		} else if strings.TrimSpace(string(text)) == "STOP" { //Check if the STOP command is being called
			fmt.Println("TCP client exiting...")
			exit <- true

			unicast_send(id, addresses[id], text, 0) //Send connection back to server to force iteration through the main for loop

		}

	}

}

func unicast_send(n int, destination string, message string, r int) {

	//Simulated message send delay
	time.Sleep(time.Duration(r) * time.Millisecond)

	//Remove trailing newline
	temp := strings.TrimSuffix(message, "\n")

	//Split for use in the print
	msg := strings.Split(temp, "@@@")

	//Opens connection to other server
	c, err := net.Dial("tcp", destination)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Outputs message through the connection
	fmt.Fprintf(c, message+"\n")

	//Time message was sent
	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"

	fmt.Println("Sent " + msg[1] + " to process " + strconv.Itoa(n) + ", system time is " + myTime)
}

func waitForMessages(ch chan string) {
	//Dedicated to checking the channel for messages passed by the server
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
	} //Ensure a process id is provided

	addresses, delays, id := readArgs(a) //Create map of process ID:"IPaddress:port", array of [min, max] delay, and current process ID

	//Extract port number in format ":XXXXX" for listener
	PORT := ":" + strings.Split(addresses[id], ":")[1]

	//More code from linode.com

	//Start listening on designated port number
	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer l.Close()

	//Create a channel for use between client and server
	ch := make(chan string)

	//Create a channel to exit the program when STOP is called
	exit := make(chan bool, 1)

	//Launch client routine
	go client(addresses, delays, id, ch, exit) //will start first client

	//Infinitely accept incoming connections and launch individual routines to handle each one.
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c, ch)
		if <-exit { //If the exit call is filled with true, the loop ends
			return
		}

	}
}

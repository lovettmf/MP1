# MP1

<h3> The System Diagram:</h3>

![Code Diagram](https://github.com/lovettmf/MP1/blob/2174778cff164f7152ab09f2e90a4267c80dfa9d/MP1%20System%20Diagram.png?raw=true)

<h3> The Code Flow: </h3>
Main function call() -> call readArgs() -> return readArgs() -> create listener server -> launch Client() goroutine -> Infinite loop -> accept messages from other processes -> launch handleConnection() goroutine -> exit when true

Client() -> launch waitForMessages() goroutine -> Inifite loop -> accept std_in, if send command launch unicast_send() goroutine. if stop kill program

unicast_send() -> dial destination server -> send message to server -> Print confirmation of send -> routine dies

handleConnection() -> read data from TCP connection -> call unicast_receive() -> close connection

waitForMessages() -> Infinite loop -> Waits for message in channel -> Prints message

<h3> Program Exeuction: </h3>
Process creation:
./go run tcp1.go [1/2/3/4]

To send a message
send [Process ID] [message]

To kill a process:
STOP

Any commands typed into the standard input without either of those two commands will be ignored.

Sending a message to an incorrect PID will result in a missing address but the program will continue to run.

package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"net"
	"log"
	"sync"
	"time"
	"encoding/json"
)

type TransactionCommand struct {
	TransactionId int
	Server string
	Command string 
	Key string
	NewValue string
	ClientName string
}

type Conn struct {
	connection net.Conn
	reader *bufio.Reader
}

type ThreadSafeMessageQueue struct {
	sync.RWMutex
	internal []string
}

func main() {
	connections := makeConnections("8888")
	messageQueue := &ThreadSafeMessageQueue{
		internal: make([]string, 0),
	}
	go readStdin(messageQueue)
	takeCommands(connections, messageQueue)
}

func readStdin(messageQueue *ThreadSafeMessageQueue) {
	reader := bufio.NewReader(os.Stdin)
	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("[-] Err reading string from STDIN: %v\n", err)
			return
		}
		messageQueue.Lock()
		messageQueue.internal = append(messageQueue.internal, command)
		messageQueue.Unlock()
	}
}

func makeSockAddr(ip string, port string) string{
	return fmt.Sprintf("%s:%s", ip, port)
}

func clientName() string {
	return os.Getenv("CLIENT_NAME")
}

func makeConn(ip string, port string) net.Conn {
	conn, err := net.Dial("tcp", makeSockAddr(ip, port))
	if err != nil {
		log.Fatal("Could not form connection to server %v\n", err)
	}

	return conn
}

func makeConnections(port string) map[string]Conn {
	servers := map[string]string{
		"A":"172.22.94.188",
		"B":"172.22.156.179",
		"C":"172.22.158.179",
		"D":"172.22.94.189",
		"E":"172.22.156.180",
	}

	connections := make(map[string]Conn)
	for server, ip := range servers {
		conn := makeConn(ip, port)
		reader := bufio.NewReader(conn)
		connection := Conn{conn, reader}
		connections[server] = connection
	}
	return connections
}

func takeCommands(connections map[string]Conn,
									messageQueue *ThreadSafeMessageQueue) {
	fmt.Printf("%d connections initialized\n", len(connections))

	inTransaction := false
	transactionId := 0
	for true {
		messageQueue.RLock()
		command := ""
		if len(messageQueue.internal) == 0 {
			messageQueue.RUnlock()
			continue
		} else {
			command = messageQueue.internal[0]
			messageQueue.RUnlock()
			messageQueue.Lock()
			messageQueue.internal = messageQueue.internal[1:]
			messageQueue.Unlock()
		}
		command = strings.Trim(command, "\n")
		if strings.HasPrefix(command, "BEGIN") {
			if inTransaction {
				fmt.Println("Began new transaction before commiting last transaction")
				continue
			}
			for server, conn := range connections {
				transactionCommand := TransactionCommand{transactionId, server, "BEGIN", "", "", clientName()}
				b, _ := json.Marshal(transactionCommand)
				fmt.Printf("%s\n", string(b))
				numBytes, err := conn.connection.Write([]byte(string(b) + "\n"))
				if err != nil {
					log.Fatal("Error writing to connection %s: %v\n", getRemoteAddr(conn.connection), err)
				}
				fmt.Println("Wrote %d bytes to %s\n", numBytes, getRemoteAddr(conn.connection))
			}
			fmt.Println("Began New Transaction")
			inTransaction = true
		} else if strings.HasPrefix(command, "GET") {
			if !inTransaction {
				fmt.Println("Did not begin a transaction")
				continue
			}
			server, key := parseGet(command)
			transactionCommand := TransactionCommand{transactionId, server, "GET", key, "", clientName()}
			b, _ := json.Marshal(transactionCommand)
			connections[server].connection.Write([]byte(string(b) + "\n"))
			fmt.Printf("%s\n", string(b))

			fmt.Printf("[+] Getting value with key %s from server %s\n", key, server)
			// resp, _ := connections[server].reader.ReadString('\n')

			input := make(chan string, 1)
			go getInput(input, connections[server].reader)
			resp := ""
			for {
        select {
        case i := <-input:
            fmt.Println("result")
            fmt.Println(i)
						resp = i
        case <-time.After(4000 * time.Millisecond):
						messageQueue.RLock()
						for idx := range messageQueue.internal {
							if strings.HasPrefix(messageQueue.internal[idx], "ABORT") {
								resp = "ABORT"
							}
						}
						messageQueue.RUnlock()
            fmt.Println("timed out")
        }
				if resp != "" {
					break
				}
    	}
			if strings.HasPrefix(resp, "ABORT") {
				if !inTransaction {
					fmt.Println("Did not start new transaction with BEGIN command")
				}
				for server, conn := range connections {
					transactionCommand := TransactionCommand{transactionId, server, "ABORT", "", "", clientName()}
					b, _ := json.Marshal(transactionCommand)
					fmt.Printf("%s\n", string(b))
					conn.connection.Write([]byte(string(b) + "\n"))
				}
				inTransaction = false
				transactionId += 1
			} else {
				fmt.Printf("%s\n", resp)
			}
		} else if strings.HasPrefix(command, "SET") {
			if !inTransaction {
				fmt.Println("Did not begin a transaction")
			}
			server, key, newValue := parseSet(command)
			transactionCommand := TransactionCommand{transactionId, server, "SET", key, newValue, clientName()}
			b, _ := json.Marshal(transactionCommand)
			connections[server].connection.Write([]byte(string(b) + "\n"))
			fmt.Printf("%s\n", string(b))

			fmt.Printf("[+] Setting value with key %s from server %s to %s\n", key, server, newValue)
			resp, _ := connections[server].reader.ReadString('\n')
			fmt.Printf("%s\n", resp)
		} else if strings.HasPrefix(command, "COMMIT") {
			if !inTransaction {
				fmt.Println("Did not start new transaction with BEGIN command")
				continue
			}

			for server, conn := range connections {
				transactionCommand := TransactionCommand{transactionId, server, "COMMIT", "", "", clientName()}
				b, _ := json.Marshal(transactionCommand)
				fmt.Printf("%s\n", string(b))
				conn.connection.Write([]byte(string(b) + "\n"))
			}

			fmt.Println("Commited transaction")
			inTransaction = false
			transactionId += 1
		} else if strings.HasPrefix(command, "ABORT") {
			if !inTransaction {
				fmt.Println("Did not start new transaction with BEGIN command")
			}
			for server, conn := range connections {
				transactionCommand := TransactionCommand{transactionId, server, "ABORT", "", "", clientName()}
				b, _ := json.Marshal(transactionCommand)
				fmt.Printf("%s\n", string(b))
				conn.connection.Write([]byte(string(b) + "\n"))
			}
			inTransaction = false
			transactionId += 1
		}
	}
}

func getInput(input chan string, reader *bufio.Reader) {
    for {
        result, err := reader.ReadString('\n')
        if err != nil {
            log.Fatal(err)
        }

        input <- result
    }
}

func parseGet(command string) (string, string){
	arr := strings.Split(command[4:], ".")
	return arr[0], arr[1]
}

func parseSet(command string) (string, string, string) {
	arr := strings.Split(command[4:], " ")
	serverValue, newValue := arr[0], arr[1]
	arr = strings.Split(serverValue, ".")
	return arr[0], arr[1], newValue
}

func getRemoteAddr(conn net.Conn) string {
	return conn.RemoteAddr().String()
}

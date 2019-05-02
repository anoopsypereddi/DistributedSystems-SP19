package main

import (
	"fmt"
	"os"
	"net"
	"log"
	"sync"
	"time"
	"bufio"
	"strings"
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

type CoordinatorCommand struct {
  Server string
	Command string 
	Key string
  Client string
}

type ThreadSafeMap struct {
	sync.RWMutex
	internal map[string]string
}

type ThreadSafeMessageQueue struct {
	sync.RWMutex
	internal []TransactionCommand
}

func serverName() string {
	return os.Getenv("SERVER_NAME")
}

func main() {
	fmt.Printf("[+] Starting server %s\n", serverName() )
	objects := &ThreadSafeMap {
		internal: make(map[string]string),
	}
  corrConn := makeConn("172.22.94.188", "4444")
  listenForClients("8888", objects, corrConn)
}

func listenForClients(port string, objects *ThreadSafeMap, corrConn net.Conn) {
	ln, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("[-] Connection from %s not accepted properly: %v\n", 
										getRemoteAddr(conn), err)
			continue
		}
		fmt.Printf("Connection formed with %s\n", getRemoteAddr(conn))
		go connectionListener(conn, objects, corrConn)
	}
}

func readConn(messageQueue *ThreadSafeMessageQueue,
							conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("[-] Err reading string from STDIN: %v\n", err)
			return
		}
		messageQueue.Lock()
		command = strings.Trim(command, "\n")
		var transactionCommand TransactionCommand
		jsonErr := json.Unmarshal([]byte(command), &transactionCommand)
		if jsonErr != nil {
			log.Fatal("[-] Error unmarshalling json: %v\n", jsonErr)
		}
		messageQueue.internal = append(messageQueue.internal, transactionCommand)
		messageQueue.Unlock()
	}
}

func connectionListener(conn net.Conn, objects *ThreadSafeMap, corrConn net.Conn) {
	messageQueue := &ThreadSafeMessageQueue{
		internal: make([]TransactionCommand, 0),
	}
	go readConn(messageQueue, conn)
	corrReader := bufio.NewReader(corrConn)
	currTransactionId := -1
	deltas := make(map[string]string)
	for {
		messageQueue.RLock()
		var transactionCommand TransactionCommand
		if len(messageQueue.internal) == 0 {
			messageQueue.RUnlock()
			continue
		} else {
			transactionCommand = messageQueue.internal[0]
			messageQueue.RUnlock()
			messageQueue.Lock()
			messageQueue.internal = messageQueue.internal[1:]
			messageQueue.Unlock()
		}

		var corrCommand CoordinatorCommand
    corrCommand.Server = transactionCommand.Server
    corrCommand.Command = transactionCommand.Command
    corrCommand.Key = transactionCommand.Key
    corrCommand.Client = transactionCommand.ClientName
    corrMessage,_ := json.Marshal(corrCommand)
		
		if transactionCommand.Command == "BEGIN" {
			for k := range deltas {
				delete(deltas, k)
			}
			currTransactionId = transactionCommand.TransactionId
			fmt.Printf("TransactionId: %d\n", currTransactionId)
		} else if transactionCommand.Command == "SET" {
			fmt.Printf("[+] Setting value of key %s to %s\n", transactionCommand.Key, transactionCommand.NewValue)
			corrConn.Write([]byte(string(corrMessage) + "\n"))
      deltas[transactionCommand.Key] = transactionCommand.NewValue
			coordMessage, err := corrReader.ReadString('\n')
      if err != nil {
        log.Fatal("[-] Coordinator Message not read correctly: %v\n", err)
      }
			fmt.Printf("Coordinator Message: %s\n", coordMessage)
			if strings.HasPrefix(coordMessage, "OK") {
				conn.Write([]byte("OK\n"))
			} else if strings.HasPrefix(coordMessage, "WAIT") {
				for {
					time.Sleep(time.Second)
					corrConn.Write([]byte(string(corrMessage) + "\n"))
					coordMessage, err = corrReader.ReadString('\n')
          if err != nil {
            log.Fatal("[-] Coordinator Message not read correctly: %v\n", err)
          }
					fmt.Printf("Coordinator Message: %s\n", coordMessage)
					if strings.HasPrefix(coordMessage, "OK") {
						conn.Write([]byte("OK\n"))
						break
					} else if strings.HasPrefix(coordMessage, "WAIT") {
						continue	
					}
				}
			}
		} else if transactionCommand.Command == "GET" {
			if v, ok := deltas[transactionCommand.Key]; ok {
				fmt.Printf("Retieved Value %s\n", v)
				corrConn.Write([]byte(string(corrMessage) + "\n"))
				corrReader.ReadString('\n')
        conn.Write([]byte(fmt.Sprintf("%s.%s = %s\n", serverName(), transactionCommand.Key, v)))
			} else {
				corrConn.Write([]byte(string(corrMessage) + "\n"))
				coordMessage, err := corrReader.ReadString('\n')
        if err != nil {
          log.Fatal("[-] Coordinator Message not read correctly: %v\n", err)
        }
				fmt.Printf("Coordinator Message: %s\n", coordMessage)
				if strings.HasPrefix(coordMessage, "OK") {
					objects.RLock()
					v, ok := objects.internal[transactionCommand.Key]
					objects.RUnlock()
					if !ok {
						conn.Write([]byte("NOT FOUND\n"))
					} else {
						fmt.Printf("Key %s has value %s\n", transactionCommand.Key, v)
						conn.Write([]byte(fmt.Sprintf("%s.%s = %s\n", serverName(), transactionCommand.Key, v)))
					}
				} else if strings.HasPrefix(coordMessage, "WAIT") {
					for {
						time.Sleep(time.Second)
						corrConn.Write([]byte(string(corrMessage) + "\n"))
						coordMessage, err = corrReader.ReadString('\n')
            if err != nil {
              log.Fatal("[-] Coordinator Message not read correctly: %v\n", err)
            }
						fmt.Printf("Coordinator Message: %s\n", coordMessage)
						if strings.HasPrefix(coordMessage, "OK") {
							objects.RLock()
							v, ok := objects.internal[transactionCommand.Key]
							objects.RUnlock()
							if !ok {
								conn.Write([]byte("NOT FOUND\n"))
							} else {
								fmt.Printf("Key %s has value %s\n", transactionCommand.Key, v)
								conn.Write([]byte(fmt.Sprintf("%s.%s = %s\n", serverName(), transactionCommand.Key, v)))
							}
							break
						} else if strings.HasPrefix(coordMessage, "WAIT") {
							continue
						}
					}
				} else if strings.HasPrefix(coordMessage, "ABORT") {
					conn.Write([]byte("ABORT\n"))
					for k, _ := range deltas {
						delete(deltas, k)
					}

				}
			}
		} else if transactionCommand.Command == "COMMIT" {
			corrConn.Write([]byte(string(corrMessage) + "\n"))
      objects.Lock()
			for key, value := range deltas {
				objects.internal[key] = value
				delete(deltas, key)
			}
      objects.Unlock()
		} else if transactionCommand.Command == "ABORT" {
			corrConn.Write([]byte(string(corrMessage) + "\n"))
			for k, _ := range deltas {
				delete(deltas, k)
			}
		}
	}
}

func getRemoteAddr(conn net.Conn) string {
	return conn.RemoteAddr().String()
}

func makeConn(ip string, port string) net.Conn {
	conn, err := net.Dial("tcp", makeSockAddr(ip, port))
	if err != nil {
		log.Fatal("Could not form connection to coordinator %v\n", err)
	}
	return conn
}

func makeSockAddr(ip string, port string) string{
	return fmt.Sprintf("%s:%s", ip, port)
}

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"log"
	"strings"
	"time"
	"encoding/json"
	"sync"
)

type Listener struct {
	conn net.Conn
	name string
}

type Message struct {
	Name string `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Message string `json:"message"`
}

var set = make(map[string]bool)
var mutex = &sync.Mutex{}

func main() {
	var connections []net.Conn
	var listeners []Listener
	name, port, num_connections := parseArgs()
	num_connections -= 1


	c := make(chan Listener)
	ln, _ := net.Listen("tcp", ":" +port)
	for i := 0; i < num_connections; i++ {
		go listen(ln, c)
	}

	nodes := getOutboundNodes()

	connections_made := 0
	for {
		if connections_made == num_connections{
			break
		}
		for ip, _ := range nodes{
			conn, err := net.Dial("tcp", makeSockAddr(ip, port))
			if err == nil {
				log.Println("[+] Connection established %q", conn.RemoteAddr());
				connections = append(connections, conn)
				fmt.Fprintf(conn, name+"\n")
				connections_made += 1
				delete(nodes, ip)
				if connections_made == num_connections {
					break
				}
			}
		}
	}
	for i := 0; i < num_connections; i++ {
		listener := <-c
		listeners = append(listeners, listener)
		log.Println(listener.name)
	}

	logListeners(listeners)
	logConnections(connections)

	spawnReaders(listeners, connections)

	mainLoop(connections, name)
}

/**
 * Opens a file and reutns the file descriptor associated 
 * with the file
 * 
 * Input: filename
 * Output: file pointer
 */
func openFile(filename string) *os.File {
	fd, err := os.OpenFile(filename,
							os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[-] error opening file: %v", err)
	}
	defer fd.Close()

	return fd
}

/**
 * Sanitizes and parses command line arguements
 *
 * Input: None
 * Output: (name, port, num_nodes)
 */
func parseArgs() (string, string, int) {
	var name string
	var port string
	var num_nodes string

	if len(os.Args) >= 4 {
		name = os.Args[1]
		port = os.Args[2]
		num_nodes = os.Args[3]

		if len(os.Args) < 5 {
			log_file := openFile("mp1.log")
			log.SetOutput(log_file)
		} else if len(os.Args) == 5 && os.Args[4] == "-v"{
			log.Println("[+] Verbose output")
		} else {
			printUsage()
			os.Exit(3)
		}
	}else {
        printUsage()
        os.Exit(3)
    }

	num_connections, err := strconv.Atoi(num_nodes)
	if err != nil {
		log.Fatalf("[-] Invalid number of nodes, requires integer")
		printUsage()
		os.Exit(3)
	}
	return name, port, num_connections
}

/**
 * Combines ip address and port to form socket address
 * which needs to be used to make connectionss
 * Input: ip (Internet Protocol Address) stirng
 *				port string
 * Output: socket address string 
 */
func makeSockAddr(ip string, port string) string{
	return fmt.Sprintf("%s:%s", ip, port)
}

/**
 * Returns map of all outbound nodes in the cluster.
 * Does not include the current node's ip address in the list
 *
 */
func getOutboundNodes() map[string]bool {
	nodes := map[string]bool{
	"172.22.94.188":true,
	"172.22.156.179":true,
	"172.22.158.179":true,
	"172.22.94.189":true,
	"172.22.156.180":true,
	"172.22.158.180":true,
	"172.22.94.190":true,
	"172.22.156.181":true,
	"172.22.158.181":true,
	"172.22.94.191":true}
	delete(nodes, GetOutboundIP().String())

	return nodes
}

func logListeners(listeners []Listener) {
	for _,listener := range listeners {
		log.Println("Listeners:", listener.conn.RemoteAddr())
	}
}

func logConnections(connections []net.Conn) {
	for _,elem := range connections {
		log.Println("Connections:", elem.RemoteAddr())
	}
}

/**
 * Spawns seperate threads for each outbound node in the cluster
 * Channel will be made which returns the return code when node 
 * fails or is disconnected from the cluster
 * Prints out all incoming messages from nodes to stdout
 *
 * Input: listeners and connections
 * Output: none
 */
func spawnReaders(listeners []Listener, connections []net.Conn) {
	a := make(chan int)
	for _, listener := range listeners {
			go readerThread(listener, a, connections)
	}
}

/**
 * Seperate thread for each node in the cluster. 
 * When a node goes up, one of the listen threads will bind to 
 * the connection and return the connection to the main thread
 *
 * Input: bind to a port, channel to return listener
 * Output: None
 */
func listen(ln net.Listener, c chan Listener) {
	conn, err := ln.Accept()
	if err == nil {
		log.Println("connection made %d : %q ", conn.RemoteAddr())
		name, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Printf("[+] Name was not correctly read for ip %q\n", 
										conn.RemoteAddr())
		}
		res := new(Listener)
		res.conn = conn
		res.name = strings.Trim(name, "\n")
		c <- *res
	}
}

/**
 * Single thread mapped to each node in the cluster.
 * Listens for messages which are recieved in the form of bytes making up a json 
 * the json is turned into a message object and the json bytes are checked against a set 
 * if the bytes exist in the set nothing happens
 * if the bytes are not in the set the string is added to the set and the message is multicasted
 * after that the message is printed to stdout
 *
 * Remains in loop until connection is lost then thread exits.
 * 
 * Input: Listener (contains information about connection to node)
 *				channel (returns exit code of thread)
 *               connections (used for multicasting)
 * Output: None
 */
func readerThread(listener Listener, c chan int, connections []net.Conn) {
	for {
		jsonBlob, err := bufio.NewReader(listener.conn).ReadString('\n')
		if err != nil { // End of the file
			fmt.Println(listener.name + " has left")
			c <- 1
			break
		}
		var message Message
		json.Unmarshal([]byte(jsonBlob), &message);
		log.Println(jsonBlob)
		mutex.Lock()
		if _, ok := set[jsonBlob]; ok {
		}else {
			set[jsonBlob] = false
			multicast(connections, message)
			fmt.Print(string(message.Name) + ": " + string(message.Message))
		}
		mutex.Unlock()
	}
}

/**
 * B-multicast structure from UIUC CS425 leture 5. multicasts 
 * message to all nodes. Connection under TCP so reliable delivery
 * can be assumed. Messages which are delivered to a node which 
 * is not responsive will not be delivered. Maintains structure 
 * among different nodes in cluster. Messages are delivered in the
 * form of a json bytes string. These are decoded after sending.
 *
 * Input: connections (list of connections to outbound nodes in cluster)
 *				text (message which user wants to deliver)
 * Output: none
 */
func multicast(connections []net.Conn, message Message) {
	jsonBlob, _ := json.Marshal(message)
	set[string(jsonBlob) + "\n"] = false
	for _,elem := range connections{
		fmt.Fprintf(elem, string(jsonBlob) + "\n")
	}
}

/**
 * Forever loop for main thread. Reads from stdin and multicasts
 * messages after user inputs an enter which is the delimination
 * character. 
 *
 * Input: connections (list of connections to outbound nodes in cluster)
 * Outut: none
 */
func mainLoop(connections []net.Conn, name string) {
	fmt.Println("READY")
	for {
		reader := bufio.NewReader(os.Stdin)
		text,_ := reader.ReadString('\n')
		message := Message{
			Name: name,
			Message: text,
			Timestamp: time.Now(),
		}
		multicast(connections, message)
	}
}

/**
 * Pings google ip address and retrieves the local ip address
 * of the computer making the request. 
 * Used to prevent nodes from self-connecting
 *
 * Input: None
 * Output: ip (local ip address of computer)
 */
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func printUsage() {
	fmt.Println("./mp1 <name> <port> <num_nodes> [options]")
	fmt.Println("\toptions:")
	fmt.Println("\t\t-v\tprint debug messages")
}

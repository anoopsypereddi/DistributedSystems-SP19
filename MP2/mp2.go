package main

import (
	"fmt"
	"bufio"
	"log"
	"net"
	"sync"
	"os"
	"strings"
	"time"
	//"strconv"
	//"bytes"
	"encoding/json"
	"math/rand"
)

type ThreadSafeAccountMap struct {
  sync.RWMutex
  internal map[int]int
}

type ThreadSafeStringToConnMap struct {
	sync.RWMutex
	internal map[string]net.Conn
}

type ThreadSafeNodeMap struct {
	sync.RWMutex
	internal map[string]Node
}

type ThreadSafeTransactionMap struct {
	sync.RWMutex
	internal map[string]Transaction
}

type ThreadSafeTransactionArr struct {
  sync.RWMutex
  internal []Transaction
}

type ThreadSafeFile struct {
	sync.RWMutex
	internal *bufio.Writer
}

type Node struct {
	name	string
	ip		string
	port	string
	alive bool
}

type Transaction struct {
	TransactionId	string `json:"transactionId"`
	Src						int `json:"src"`
	Dest					int `json:"dest"`
	Amount				int `json:"amount"`
}

func main() {
	rand.Seed(time.Now().Unix())
	name, serviceIp, servicePort := parseArgs()
	localIp := GetOutboundIP()
	localPort := getFreePort()

	log.Printf("[+] Running on %s:%s\n", localIp, localPort)

	connections := &ThreadSafeStringToConnMap {
		internal: make(map[string]net.Conn),
	}

	nodesMap := &ThreadSafeNodeMap {
		internal: make(map[string]Node),
	}

	transactions := &ThreadSafeTransactionMap {
		internal: make(map[string]Transaction),
	}

  transactionArr := &ThreadSafeTransactionArr {
    internal: make([]Transaction, 0),
  }

	transactions_fd := &ThreadSafeFile {
		internal: bufio.NewWriter(openFile(fmt.Sprintf("logs/%s.transactions", name))),
	}

	blockchain := initializeBlockchain()

<<<<<<< HEAD
	serviceConn := connectToService(name, serviceIp, servicePort, localIp, localPort)
	go manageConnections(name, serviceConn, nodesMap, connections, transactions, transactions_fd, transactionArr, blockchain)
=======
	go manageConnections(name, nodesMap, connections, transactions, transactions_fd, transactionArr)
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c

	nodesMap.Lock()
	nodesMap.internal[name] = parseIntro(introductionMessage(name,
																												serviceIp,
																												servicePort))
	nodesMap.Unlock()

<<<<<<< HEAD
	go listenOnPort(name, serviceConn, localIp, localPort, connections, nodesMap, transactions, transactions_fd, transactionArr, blockchain)
	go takeCommands(name, serviceConn, nodesMap, connections, transactions, transactionArr, blockchain)

=======
	go listenOnPort(name, localIp, localPort, connections, nodesMap, transactions, transactions_fd, transactionArr)
	go takeCommands(name, nodesMap, connections, transactions, blockchain)

	serviceConn := connectToService(name, serviceIp, servicePort, localIp, localPort)
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
	listenToService(name, localIp, localPort, serviceConn, nodesMap, connections, transactions, transactions_fd, transactionArr, blockchain)
}


/**
 * Go routine which continuously runs. Every second this routine will check if the number
 * of connections is greater than or equal to the sqrt of the number of nodes. If it is not
 * then it will find a new connection and form a new connection listener and will start 
 * listening to another node. 
 *
 */
func manageConnections(	name string,
												serviceConn net.Conn,
												nodesMap *ThreadSafeNodeMap,
												connections *ThreadSafeStringToConnMap,
												transactions *ThreadSafeTransactionMap,
												transactions_fd *ThreadSafeFile,
                        transactionArr *ThreadSafeTransactionArr,
												blockchain *Blockchain) {
	time.Sleep(2 * time.Second)
	for {
		time.Sleep(time.Second)
		nodesMap.RLock()
		numNodes := len(nodesMap.internal)
		nodesMap.RUnlock()

		connections.RLock()
		numConnections := len(connections.internal)
		connections.RUnlock()

		for numConnections * numConnections < numNodes && numNodes > 2{
			_, _, err := findNewConnection(connections, nodesMap, transactions)
			if err {
				log.Printf("[-] Could not form a new connection")
			} else {
				//go connectionListener(name, serviceConn, conn, connections, nodesMap, reader, transactions, transactions_fd, transactionArr, blockchain)
			}
			numConnections += 1
		}
	}
}

/**
 * Takes commands from stdin and executes. For 
 * Testing purpouses only.
 *
 *
 *
 */
func takeCommands(name string,
									serviceConn net.Conn,
									nodesMap *ThreadSafeNodeMap,
									connections *ThreadSafeStringToConnMap,
<<<<<<< HEAD
									transactions *ThreadSafeTransactionMap,
=======
									transactions *ThreadSafeTransactionSet,
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
                  transactionArr *ThreadSafeTransactionArr,
									blockchain *Blockchain) {
	reader := bufio.NewReader(os.Stdin)
	for {
		message, err := reader.ReadString('\n')

		if err != nil {
			log.Printf("[+] Error reading from stdin: %v\n", err)
		}
		if strings.HasPrefix(message, "LIST CONNECTIONS") {
			connections.RLock()
			for _, v := range connections.internal {
				fmt.Printf("%s <-> %s\n",
											v.RemoteAddr().String(),
											v.LocalAddr().String())
			}
			connections.RUnlock()
		} else if strings.HasPrefix(message, "LIST NODES") {

			nodesMap.RLock()
			for name, node := range nodesMap.internal {
				log.Printf("%s => %s:%s", name, node.ip, node.port)
			}
			nodesMap.RUnlock()

		} else if strings.HasPrefix(message, "LIST NAMES") {
			connections.RLock()
			for name, conn := range connections.internal {
				log.Printf("%s => %s", name, getRemoteAddr(conn))
			}
			connections.RUnlock()
		} else if strings.HasPrefix(message, "GENTRANS") {
			arr := strings.Fields(message)
			transaction := parseTransaction(generateTransaction(arr[1]))

			transactions.Lock()
			transactions.internal[transaction.TransactionId] = transaction
			transactions.Unlock()
      insertTransaction(transaction, transactionArr)

			multicast(connections, generateTransaction(arr[1]))
		} else if strings.HasPrefix(message, "LIST TRANSACTIONS") {
			transactionArr.RLock()
			for i := range transactionArr.internal {
				log.Printf("[TRANSACTION] %s\n", transactionArr.internal[i].TransactionId)
			}
			transactionArr.RUnlock()
		} else if strings.HasPrefix(message, "SNAPSHOT") {
			arr := strings.Fields(message)
			snapshotName := fmt.Sprintf("logs/%s-%s.snapshot", name, arr[1])
			fd, _ := os.Create(snapshotName)
			writer := bufio.NewWriter(fd)
			connections.RLock()
			for nodeName, _ := range connections.internal {
				writer.WriteString(fmt.Sprintf("%s -> %s\n", name, nodeName))
			}
			writer.Flush()
			connections.RUnlock()
<<<<<<< HEAD
		} else if strings.HasPrefix(message, "CURR_BLOCK") {
			log.Printf("PREV_HASH -> %s CURR_HASH -> %s DEPTH -> %d\n", blockchain.currentBlock.PrevHash,
																																	blockchain.currentBlock.Hash,
																																	blockchain.currentBlock.Depth)
		} else if strings.HasPrefix(message, "SOLVING_BLOCK") {
			log.Printf("PREV_HASH -> %s CURR_HASH -> %s DEPTH -> %d\n", blockchain.solvingBlock.PrevHash,
																																	blockchain.solvingBlock.Hash,
																																	blockchain.solvingBlock.Depth)

		} else if strings.HasPrefix(message, "MINE_CURR_BLOCK") {
			moveCurrBlockToSolveBlock(blockchain, serviceConn)
		} else if strings.HasPrefix(message, "LIST BLOCKS") {
			for _, block := range blockchain.blocks {
				log.Printf("\t|%s\n", block.Hash)
			}
		} else if strings.HasPrefix(message, "LAST HASH") {
			blockchain.Lock()
			log.Printf("%s\n", blockchain.longestChainHash)
			blockchain.Unlock()
			/*
			transactions.Lock()
			log.Printf("transactions\n")
			transactions.Unlock()
			transactionArr.Lock()
			log.Printf("transactionsArr\n")
			transactionArr.Unlock()
			nodesMap.Lock()
			log.Printf("nodesMap\n")
			nodesMap.Unlock()
			connections.Lock()
			log.Printf("connections\n")
			connections.Unlock()
			*/
		} else if strings.HasPrefix(message, "BLOCK_TRANSACTIONS") {
			arr := strings.Fields(message)
			if len(arr) != 2 {
				log.Printf("Invalid Command")
				continue
			}
			if block, ok := blockchain.blocks[arr[1]]; ok {
				for _, transaction := range block.Transactions {
					log.Printf("%s\n", transaction.String())
				}
			} else if block, ok := blockchain.unverifiedBlocks[arr[1]]; ok {
				for _, transaction := range block.Transactions {
					log.Printf("%s\n", transaction.String())
				}
			} else {
				log.Printf("[+] Block %s not found in blockchain or unverified queue",
											arr[1])
			}
		} else if strings.HasPrefix(message, "ACCOUNT_BALANCES") {
			arr := strings.Fields(message)
			if len(arr) != 2 {
				log.Printf("Invalid Command")
				continue
			}
			if block, ok := blockchain.blocks[arr[1]]; ok {
				for acct, bal := range block.AccountBalances {
					log.Printf("%d -> %d\n", acct, bal)
				}
			} else if block, ok := blockchain.unverifiedBlocks[arr[1]]; ok {
				for acct, bal := range block.AccountBalances {
					log.Printf("%d -> %d\n", acct, bal)
				}
			} else {
				log.Printf("[+] Block %s not found in blockchain or unverified queue",
											arr[1])
			}
=======
		} else if strings.HasPrefix(message, "MINE_CURR_BLOCK") {
			moveCurrBlockToSolveBlock(blockchain, serviceConn)
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
		}

	}
}

/**
 * Continuously listens for commands from the service. These 
 * can include 
 *	INTRODUCE <node_name> <node_ip> <node_port>
 *	TRANSACTION <time> <transaction_id> <to> <from> <amount>
 *	QUIT
 *	DIE
 */
func listenToService(	name string,
											localIp string,
											localPort string,
											serviceConn net.Conn,
											nodesMap *ThreadSafeNodeMap,
											connections *ThreadSafeStringToConnMap,
											transactions *ThreadSafeTransactionMap,
											transactions_fd *ThreadSafeFile,
                      transactionArr *ThreadSafeTransactionArr,
<<<<<<< HEAD
                      blockchain *Blockchain) {
	reader := bufio.NewReader(serviceConn)
=======
											blockchain *Blockchain) {
	reader := bufio.NewReader(conn)
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("[-] Service Connection Lost")
			time.Sleep(3 * time.Second)
			os.Exit(2)
		}
		if strings.HasPrefix(message, "INTRODUCE") {
			// TODO: Gossip new node intro
			multicast(connections, message)

			// Adding node to known nodes
			node := parseIntro(message)

			nodesMap.RLock()
			_, ok := nodesMap.internal[node.name]
			nodesMap.RUnlock()

			if !ok {
				nodesMap.Lock()
				nodesMap.internal[node.name] = node
				nodesMap.Unlock()
			}
      log.Printf(message)
			// Forming connection to node
			conn, err := net.Dial("tcp", makeSockAddr(node.ip, node.port))
			if err != nil {
				log.Printf("[-] Could not make connection to %s:%s, error %v\n",
										node.ip, node.port, err)
				continue
			} else {
        connections.Lock()
        connections.internal[name] = conn
        connections.Unlock()
        
        
        nodesMap.RLock()
        for _, curr_node := range nodesMap.internal {
          writeToConn(conn, introductionMessage(curr_node.name, curr_node.ip, curr_node.port))
        }
        nodesMap.RUnlock()

        transactions.RLock()
        for _, transaction := range transactions.internal {
          writeToConn(conn, transactionMessage(transaction))
        }
        transactions.RUnlock()
        
      }

			//reader := bufio.NewReader(conn)

			writeToConn(conn, nameMessage(name))
			writeToConn(conn, introductionMessage(name, localIp, localPort))


			// log.Printf("[+] Connected(Introduce) to %s %s:%s, %s\n", node.name, node.ip, node.port, getRemoteAddr(conn))
			//connections.Lock()
			//connections.internal[node.name] = conn
			//connections.Unlock()

<<<<<<< HEAD
			//go connectionListener(node.name, serviceConn, conn, connections, nodesMap, reader, transactions, transactions_fd, transactionArr, blockchain)
=======
			go connectionListener(node.name, conn, connections, nodesMap,
														reader, transactions, transactions_fd, transactionArr)
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c

		} else if strings.HasPrefix(message, "TRANSACTION") {
				transaction := parseTransaction(message)
				transactions.Lock()
				transactions.internal[transaction.TransactionId] = transaction
				transactions.Unlock()
        insertTransaction(transaction, transactionArr)

				multicast(connections, message)
				// log.Printf("%s\n", transaction.transactionId)

				transactions_fd.Lock()
<<<<<<< HEAD
				transactions_fd.internal.Write([]byte(fmt.Sprintf("%s\n",
																											transaction.TransactionId)))
=======
				transactions_fd.internal.Write([]byte(fmt.Sprintf("%f %s\n",
																								transaction.timestamp,
																								transaction.transactionId)))
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
				transactions_fd.Unlock()
		} else if strings.HasPrefix(message, "SOLVED") {
			log.Printf("%s\n", message)
      blockHash, solutionHash := parseSolved(message)
			minedBlock := solutionForSolving(blockchain, blockHash, solutionHash)
			multicast(connections, blockMessage(minedBlock))
<<<<<<< HEAD
		} else if strings.HasPrefix(message, "VERIFY") {
			passed, hash, _ := parseVerify(message)
			if passed {
				// log.Printf("[+] Verified block %s\n", hash)
				blockVerified(blockchain, hash)
			}
    }
=======
		}
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
	}
}

/**
 * Thread constantly listening on a specified port, 
 * any connections made to this process will be accepted
 * and if the connection is not already in the connections
 * map we have, it will be added to the connections map. 
 */
func listenOnPort(nodeName string,
									serviceConn net.Conn,
									localIp string,
									localPort string,
									connections *ThreadSafeStringToConnMap,
									nodesMap *ThreadSafeNodeMap,
									transactions *ThreadSafeTransactionMap,
									transactions_fd *ThreadSafeFile,
                  transactionArr *ThreadSafeTransactionArr,
									blockchain *Blockchain) {
	ln, err := net.Listen("tcp", ":" + localPort)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Printf("[-] Connection from %s not accepted properly\n",
										getRemoteAddr(conn))
		} else {
			writeToConn(conn, nameMessage(nodeName))
			reader := bufio.NewReader(conn)
			message, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("[-] Error reading name from %s\n", getRemoteAddr(conn))
				continue
			}
			name := parseName(message)

			if nodeInConnections(name, connections) {
				// log.Printf("[-] Connection to %s already in connections.", getRemoteAddr(conn))
        conn.Close()
			} else {
				// log.Printf("[+] Connection(Listener) from %s accepted\n", getRemoteAddr(conn))
				go connectionListener(name, serviceConn, conn, connections, nodesMap, reader, transactions, transactions_fd, transactionArr, blockchain)
			}
		}
	}
}

/**
 * Each connection will be mapped to a single thread. Each listener
 * thread will take messages from the node it is mapped to. 
 * Connections will be dropped when there is an EOF error
 * from the reader listening to that node. This method can parse INTRODUCE, 
 * TRANSACTION, DELETE messages. 
 * INTRODUCE <name> <ip> <port>
 * TRANSACTION <timestamp> <transactionId> <sourceId> <destId> <amount>
 * DELETE <name>
 * 
 * TODO: When a node is deleted attempt to form new connections
 * TODO: There should be 5 consecutive connections at one time
 */
func connectionListener(name string,
												serviceConn net.Conn,
												conn net.Conn,
												connections *ThreadSafeStringToConnMap,
												nodesMap *ThreadSafeNodeMap,
												reader *bufio.Reader,
												transactions *ThreadSafeTransactionMap,
												transactions_fd *ThreadSafeFile,
<<<<<<< HEAD
                        transactionArr *ThreadSafeTransactionArr,
												blockchain *Blockchain) {
  
=======
                        transactionArr *ThreadSafeTransactionArr) {
	connections.Lock()
	connections.internal[name] = conn
	connections.Unlock()

	nodesMap.RLock()
	for _, curr_node := range nodesMap.internal {
		writeToConn(conn, introductionMessage(curr_node.name,
																					curr_node.ip,
																					curr_node.port))
	}
	nodesMap.RUnlock()

	transactions.RLock()
	for transaction, _ := range transactions.internal {
		writeToConn(conn, transactionMessage(transaction))
	}
	transactions.RUnlock()
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
<<<<<<< HEAD
			log.Printf("[DISCONNECT] %v", err)
      
      connections.Lock()
      delete(connections.internal, name)
=======
			log.Printf("[DISCONNECT] %s <-> %s",	getRemoteAddr(conn),
																						conn.LocalAddr().String())
			connections.Lock()
			delete(connections.internal, name)
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
			connections.Unlock()
      
      nodesMap.RLock()
      node := nodesMap.internal[name] 
      nodesMap.RUnlock() 
      //log.Printf("WriteConnection - Dial: %s", makeSockAddr(node.ip, node.port))
      wConn, err := net.Dial("tcp", makeSockAddr(node.ip, node.port))
      if(err != nil) {
        multicast(connections, deleteMessage(name))
        //log.Printf("WriteConnection - Fail")
      } else {
        log.Printf("[RECONNECT] %v", err)
        connections.Lock()
        connections.internal[name] = wConn
        connections.Unlock()
      }
      return 
		} else if strings.HasPrefix(message, "INTRODUCE") {
			// Adding node to known nodes
			node := parseIntro(message)

			nodesMap.RLock()
			_, ok := nodesMap.internal[node.name]
			nodesMap.RUnlock()
      
			if ok {
				continue
			} else {
				nodesMap.Lock()
				nodesMap.internal[node.name] = node
				nodesMap.Unlock()
				multicast(connections, message)
			}
      if node.name == name {
        createWriteConnection(nodesMap, name, connections, transactions)
      }
		} else if strings.HasPrefix(message, "DELETE") {
			deleteNodeName := parseName(message)

			nodesMap.Lock()
			if _ , ok := nodesMap.internal[deleteNodeName]; ok {
				delete(nodesMap.internal, deleteNodeName)
				multicast(connections, message)
			}
			nodesMap.Unlock()

		} else if strings.HasPrefix(message, "TRANSACTION") {
			transaction := parseTransaction(message)
			//log.Printf("[+] Recieved transaction %s\n", transaction.transactionId)
			transactions.RLock()
			_, ok := transactions.internal[transaction.TransactionId]
			transactions.RUnlock()
			if ok {
				// log.Printf("[+] Transaction %s already seen\n", transaction.transactionId)
				continue
			} else {
				multicast(connections, message)
				transactions.Lock()
				transactions.internal[transaction.TransactionId] = transaction
				transactions.Unlock()
				insertTransaction(transaction, transactionArr)
<<<<<<< HEAD
				insertTransactionIntoBlock(blockchain, &transaction)
				transactions_fd.Lock()
				transactions_fd.internal.Write([]byte(fmt.Sprintf("%s\n",
																											transaction.TransactionId)))
=======

				transactions_fd.Lock()
				transactions_fd.internal.Write([]byte(fmt.Sprintf("%f %s\n",
																					transaction.timestamp,
																					transaction.transactionId)))
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
				transactions_fd.internal.Flush()
				transactions_fd.Unlock()
			}
		} else if strings.HasPrefix(message, "BLOCK_SOLVED") {
<<<<<<< HEAD
			arr := strings.Fields(message)
		  jsonBlob := []byte(arr[1])	
			//log.Printf("Recieved block %s\n", jsonBlob)
			block := Block{}
			err := json.Unmarshal(jsonBlob, &block)
			if err != nil {
				log.Printf("Error unmarshalling data %v\n", err)
			}
			//log.Printf("Hash: %s\n", block.Hash)
=======
			jsonBlob := strings.TrimSpace(message[12:])
			block := Block{}
			json.Unmarshal([]byte(jsonBlob), &block)
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
			blockchain.Lock()
			_, inBlockchain := blockchain.blocks[block.Hash]
			_, inUnverified := blockchain.unverifiedBlocks[block.Hash]

			if inBlockchain || inUnverified {
				blockchain.Unlock()
				continue
			}
<<<<<<< HEAD
			blockchain.unverifiedBlocks[block.Hash] = &block
			blockchain.Unlock()
			multicast(connections, blockMessage(&block))
			writeToConn(serviceConn, verifyCommand(blockHash(&block), block.SolvedHash))
			//log.Printf("[+] Verifying Block %s\n", block.Hash)
=======
			multicast(connections, message)
			blockchain.unverifiedBlocks[block.Hash] = &block
			blockchain.Unlock()
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
		}
	}
}

func createWriteConnection(nodesMap     *ThreadSafeNodeMap,
                           name         string,
                           connections  *ThreadSafeStringToConnMap,
                           transactions *ThreadSafeTransactionMap) {
  nodesMap.RLock()
  node := nodesMap.internal[name] 
  nodesMap.RUnlock() 
  //log.Printf("WriteConnection - Dial: %s", makeSockAddr(node.ip, node.port))
  wConn, err := net.Dial("tcp", makeSockAddr(node.ip, node.port))
  if(err != nil) {
    //log.Printf("WriteConnection - Fail")
  } else {
	  connections.Lock()
	  connections.internal[name] = wConn
	  connections.Unlock()
	  nodesMap.RLock()
    for _, curr_node := range nodesMap.internal {
		  writeToConn(wConn, introductionMessage(curr_node.name, curr_node.ip, curr_node.port))
  	}
    nodesMap.RUnlock()

	  transactions.RLock()
	  for _, transaction := range transactions.internal {
		  writeToConn(wConn, transactionMessage(transaction))
	  }
	  transactions.RUnlock()
  }
}

/**
 * Iterates through all of the nodes inside of the network and 
 * attempt to make connections with them. Iterating through a map
 * will automatically randomize order and will prevent (statistically)
 * which will prevent 
 *
 */
func findNewConnection(	connections *ThreadSafeStringToConnMap,
<<<<<<< HEAD
												nodesMap *ThreadSafeNodeMap,
                        transactions *ThreadSafeTransactionMap) (net.Conn, string, bool) {
=======
												nodesMap *ThreadSafeNodeMap) (net.Conn,
																											*bufio.Reader,
																											string,
																											bool) {
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
	nodesMap.Lock()
	defer nodesMap.Unlock()
	for _ , node := range nodesMap.internal {
		connections.RLock()
		_, ok := connections.internal[node.name]
		connections.RUnlock()

		if ok {
			continue
		} else {
			log.Printf("[+] Attempting connection to %s\n", node.name)
			conn, err := net.Dial("tcp", makeSockAddr(node.ip, node.port))
			if err != nil {
				log.Printf("[-] Could not connect to %s\n", node.name)
				multicast(connections, deleteMessage(node.name))
				continue
			} else {

				log.Printf("[+] Formed new connection with %s\n", node.name)

				connections.Lock()
				connections.internal[node.name] = conn
				connections.Unlock()

        for _, curr_node := range nodesMap.internal {
          writeToConn(conn, introductionMessage(curr_node.name, curr_node.ip, curr_node.port))
        }

        transactions.RLock()
        for _, transaction := range transactions.internal {
          writeToConn(conn, transactionMessage(transaction))
        }
        transactions.RUnlock()   

        return conn, node.name, true
			}
		}
	}
	return nil, "", false
}

/**
 * Makes initial connection to service
 * and writes message to the connection. 
 *
 * Format of message:
 *		CONNECT <name> <local_ip_address> <local_port>
 *
 * If connection is successful, then the conenction is 
 * returned to go into the main loop of the program.
 */
func connectToService(name string,
											serviceIp string,
											servicePort string,
											localIp string,
											localPort string) (net.Conn) {

	conn, err := net.Dial("tcp", makeSockAddr(serviceIp, servicePort))

	if err != nil {
		log.Fatal("[-] Connection to service failed")
	}

	writeToConn(conn, serviceConnectMessage(name, localIp, localPort))
	return conn
}



package main

import (
	"fmt"
	"os"
	"log"
	"net"
	"strconv"
)

/**
* Combines ip address and port to form socket address
* which needs to be used to make connectionss
* Input: ip (Internet Protocol Address) stirng
*              port string
* Output: socket address string
*/
func makeSockAddr(ip string, port string) string{
	return fmt.Sprintf("%s:%s", ip, port)
}

func parseArgs() (string, string, string){
	log.SetFlags(0)

	i := 0

	if os.Args[2] == "-v" {
		i = 1
	} else {
		// TODO: Set log output to file
	}

	// TODO: Restructure error handling
	if len(os.Args) < 4 || len(os.Args) >= 5 {
		fmt.Printf("[-] Invalid number of arguements\n")
		printUsage()
		os.Exit(1)
	} else if !isIpValid(os.Args[i + 2]) {
		fmt.Printf("[-] Invalid service ip address\n")
		printUsage()
		os.Exit(1)
	} else if !isNumber(os.Args[i + 3]) {
		fmt.Printf("[-] Invalid service port\n")
		printUsage()
		os.Exit(1)
	}

	name					:= os.Args[i + 1]
	serviceIp			:= os.Args[i + 2]
	servicePort		:= os.Args[i + 3]

	return name, serviceIp, servicePort
}

func (t Transaction) String () string {
	return fmt.Sprintf("%s %d %d %d", t.TransactionId, t.Src, t.Dest, t.Amount)
}

func insertTransaction(t Transaction, transactions *ThreadSafeTransactionArr) {
  transactions.Lock()
  defer transactions.Unlock()
<<<<<<< HEAD
	/*
  idx := len(transactions.internal) - 1
  transactions.internal = append(transactions.internal, t)
  if idx == -1 {
    return
  }
  for transactions.internal[idx + 1].Timestamp < transactions.internal[idx].Timestamp {
    transactions.internal[idx], transactions.internal[idx+1] = transactions.internal[idx+1], transactions.internal[idx]
    idx -= 1
    if idx == -1 {
      break
    }
  }
	*/
	transactions.internal = append(transactions.internal, t)
  return
=======
  transactions.internal = append(transactions.internal, t) 
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
}

func nodeInConnections(	name string,
												connections *ThreadSafeStringToConnMap) bool {
	connections.RLock()
	defer connections.RUnlock()
	_, ok := connections.internal[name]
	return ok
}

func printUsage() {
	fmt.Printf("\tUsage: mp2 [-v] <name> <service_ip> <service_port>\n")
}

func isIpValid(ip string) bool {
	addr := net.ParseIP(ip)
	return addr != nil
}

func isNumber(v string) bool {
	_, err := strconv.Atoi(v)
	return err == nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func openFile(filename string) *os.File{
	fd, err := os.Create(filename)
	if err != nil {
		log.Fatal("[-] Error while opening file")
	}
	return fd
}

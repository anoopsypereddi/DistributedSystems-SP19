package main

import (
	"net"
	"strconv"
	"log"
//	"strings"
	"fmt"
//	"time"
)

/**
 * Writes message to a net.Conn
 * Will function even if net.Conn is no longer valid
 *
 */
func writeToConn(conn net.Conn, msg string) {
	_, err := fmt.Fprintf(conn, msg) 
	if err != nil {
		return
	}
}

/**
 * Gets the remote address of a net.Conn and returns ip
 * in string form
 *
 */
func getRemoteAddr(conn net.Conn) string {
	return conn.RemoteAddr().String()
}

/**
 * Forms a connection with itself and returns the port which
 * was bound to. The port is thus open and can be used to form
 * connections
 */
func getFreePort() (string) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Fatal("[-] Cannot find a free port")
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("[-] Cannot find a free port")
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

/**
 * Pings google ip address and retrieves the local ip address
 * of the computer making the request.
 * Used to prevent nodes from self-connecting
 *
 * Input: None
 * Output: ip (local ip address of computer)
 */
func GetOutboundIP() (string) {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
       // log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String()
}

func multicast(	connections *ThreadSafeStringToConnMap,
								msg string) {
		connections.Lock()
		for _, conn := range connections.internal {
			writeToConn(conn, msg)
		}
		connections.Unlock()
}

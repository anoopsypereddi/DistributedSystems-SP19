package main

import (
	"strings"
	"strconv"
	"log"
)

func parseIntro(msg string) Node {
	arr := strings.Fields(msg)
	node := Node {
		name: arr[1],
		ip: arr[2],
		port: arr[3],
		alive: true,
	}
	return node
}
func parseName(msg string) string {
	arr := strings.Fields(msg)
	return arr[1]
}

func parseTransaction(msg string) Transaction {
	arr := strings.Fields(msg)

	//timestamp, err1	:= strconv.ParseFloat(arr[1], 64)
	source, err2		:= strconv.Atoi(arr[3])
	dest, err3			:= strconv.Atoi(arr[4])
	amount, err4		:= strconv.Atoi(arr[5])

	if	(err2 != nil) ||
			(err3 != nil) ||
			(err4 != nil) {
		log.Fatal("[-] Malformed transaction %s\n", msg)
	}

	transaction := Transaction {
		TransactionId:	arr[2],
		Src:						source,
		Dest:						dest,
		Amount:					amount,
	}
	return transaction
}

func parseSolved(msg string) (string, string) {
  arr := strings.Fields(msg)
  return arr[1], arr[2]
}

func parseVerify(msg string) (bool, string, string) {
	arr := strings.Fields(msg)
	passed := true
	if arr[1] == "FAIL" {
		passed = false
	}
	return passed, arr[2], arr[3]
}

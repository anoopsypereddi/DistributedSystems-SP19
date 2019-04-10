package main

import (
	"fmt"
	"log"
	"encoding/json"
)
func serviceConnectMessage(	name string,
														ip string,
														port string) string {
	return fmt.Sprintf("CONNECT %s %s %s\n", name, ip, port)
}

func introductionMessage( name string,
													ip string,
													port string) string {
	return fmt.Sprintf("INTRODUCE %s %s %s\n", name, ip, port)
}

func transactionMessage(transaction Transaction) string {
	return fmt.Sprintf("TRANSACTION %f %s %d %d %d\n", 1.14,
												transaction.TransactionId, transaction.Src,
												transaction.Dest, transaction.Amount)
}

func blockMessage(block *Block) string {
	jsonBlob, err := json.Marshal(*block)
	if err != nil {
		log.Printf("[-] Block %s could not be seialized into json string\n", block.Hash)
	}
	return fmt.Sprintf("BLOCK_SOLVED %s\n", jsonBlob)
}

func deleteMessage( name string ) string {
	return fmt.Sprintf("DELETE %s\n", name)
}

func nameMessage( name string) string {
	return fmt.Sprintf("NAME %s\n", name)
}

func generateTransaction(transactionId string) string {
	return fmt.Sprintf("TRANSACTION 1551208414.204385 %s 183 99 10\n", transactionId)
}

<<<<<<< HEAD
func solveCommand(hash string) string {
	log.Printf("SOLVE %s\n", hash)
	return fmt.Sprintf("SOLVE %s\n", hash)
}

func verifyCommand(blockHash string, solutionHash string) string {
  //log.Printf("VERIFY %s %s\n", blockHash, solutionHash)
  return fmt.Sprintf("VERIFY %s %s\n", blockHash, solutionHash)
=======
func blockMessage(block *Block) string {
	jsonBlob, err := json.Marshal(*block)
	if err != nil {
		log.Printf("[-] Block %s could not be seialized into json string\n", block.Hash)
	}
	return fmt.Sprintf("BLOCK_SOLVED %s\n", jsonBlob)
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
}

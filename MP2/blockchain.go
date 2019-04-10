package main

import (
	"sync"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"log"
	"net"
<<<<<<< HEAD
	"fmt"
	"strings"
=======
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
)

type Blockchain struct {
	sync.RWMutex
<<<<<<< HEAD
	blocks						map[string]*Block
	unverifiedBlocks	map[string]*Block
=======
	blocks						*ThreadSafeBlockMap
	unverifiedBlocks	*ThreadSafeBlockMap
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
  solvingBlock			*Block
	currentBlock			*Block
	longestChainHash	string
	longestChainDepth	int
<<<<<<< HEAD
	transactionsBuff		*ThreadSafeTransactionMap
=======
}

type ThreadSafeBlockMap struct {
  sync.RWMutex
  internal map[string]*Block
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
}

type Block struct {
	sync.RWMutex
	Transactions		map[string]Transaction `json:"transactions"`
	PrevHash				string `json:"prevHash"`
	Hash						string `json:"hash"`
	Depth						int `json:"depth"`
	SolvedHash			string `json:"solvedHash"`
	AccountBalances	map[int]int `json:"accountBalances"`
}

<<<<<<< HEAD
type NetworkBlock struct {
	Transactions		[]string `json:"transactions"`
	PrevHash				string `json:"prevHash"`
	Hash						string `json:"hash"`
	Depth						int `json:"depth"`
	SolvedHash			string `json:"solvedHash"`
	AccountBalances	[]int `json:"accountBalances"`
}

=======
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
type SerialiazableBlock struct {
	Transactions		map[string]Transaction `json:"transactions"`
	PrevHash				string `json:"prevHash"`
	AccountBalances map[int]int `json:"accountBalances"`
}

func initBlock(prevHash string, depth int) *Block{
	return &Block {
		Transactions: make(map[string]Transaction),
		PrevHash: prevHash,
		Hash: "",
		SolvedHash: "",
		Depth: depth,
	}
}

func initializeBlockchain() *Blockchain {
	blockchain := &Blockchain {
<<<<<<< HEAD
		blocks: make(map[string]*Block),
		unverifiedBlocks: make(map[string]*Block),
=======
		blocks: &ThreadSafeBlockMap {
			internal: make(map[string]*Block)
		},
		unverifiedBlocks: &ThreadSafeBlockMap {
			internal: make(map[string]*Block)
		},
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
		solvingBlock: nil,
		currentBlock: nil,
		longestChainHash: "0",
		longestChainDepth: 0,
<<<<<<< HEAD
		transactionsBuff: &ThreadSafeTransactionMap {
			internal: make(map[string]Transaction),
		},
=======
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
	}
	blockchain.currentBlock = initBlock("0", 0)
	return blockchain
}

func moveCurrBlockToSolveBlock(blockchain *Blockchain, serviceConn net.Conn) {
<<<<<<< HEAD
	blockchain.transactionsBuff.Lock()
	for _, transaction := range blockchain.transactionsBuff.internal {
		blockchain.currentBlock.Transactions[transaction.TransactionId] = transaction
		delete(blockchain.transactionsBuff.internal, transaction.TransactionId)
	}

	blockchain.transactionsBuff.Unlock()
	blockchain.currentBlock.PrevHash = blockchain.longestChainHash
	blockchain.currentBlock.Depth = blockchain.longestChainDepth + 1

	calculateAccountBalances(blockchain)

	blockchain.currentBlock.Hash = blockHash(blockchain.currentBlock)
	log.Printf("[+] Attempting to mine block with hash %s\n", blockchain.currentBlock.Hash)
	writeToConn(serviceConn, solveCommand(blockchain.currentBlock.Hash))

	currDepth := blockchain.currentBlock.Depth
	blockchain.solvingBlock = blockchain.currentBlock
	blockchain.currentBlock = initBlock(blockchain.solvingBlock.Hash, currDepth + 1)
=======
	currDepth := blockchain.currentBlock.Depth
	blockchain.solvingBlock = blockchain.currentBlock
	blockchain.currentBlock = initBlock(blockchain.solvingBlock.Hash, currDepth + 1)

	blockchain.solvingBlock.PrevHash = blockchain.longestChainHash
	blockchain.solvingBlock.Depth = blockchain.longestChainDepth + 1

	calculateAccountBalances(blockchain)

	blockchain.solvingBlock.Hash = blockHash(blockchain.solvingBlock)
	log.Printf("[+] Attempting to mine block with hash %s\n", blockchain.solvingBlock.Hash)
	writeToConn(serviceConn, solveCommand(blockchain.solvingBlock.Hash))

>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
}

func solutionForSolving(blockchain *Blockchain, blockHash string, solutionHash string) *Block {
	if blockchain.solvingBlock.Hash != blockHash {
		log.Fatal("[-] The block which was solved was not the one in solving")
	}

	blockchain.solvingBlock.SolvedHash = solutionHash
	blockchain.blocks[blockHash] = blockchain.solvingBlock
	setLongestChain(blockchain, blockchain.solvingBlock)
	blockchain.solvingBlock = nil
	return blockchain.blocks[blockHash]
}

func calculateAccountBalances(blockchain *Blockchain) {
<<<<<<< HEAD
	block := blockchain.currentBlock
	block.AccountBalances = make(map[int]int)
=======
	blockchain.RLock()
  block := blockchain.solvingBlock
	blockcain.RUnlock()
  block.AccountBalances = make(map[int]int)

  block.Lock()
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
	if block.PrevHash != "0" {
		for acct, bal := range blockchain.blocks[block.PrevHash].AccountBalances{
			block.AccountBalances[acct] = bal
		}
<<<<<<< HEAD
=======

>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
		for hash, _ := range blockchain.blocks[block.PrevHash].Transactions {
			delete(block.Transactions, hash)
		}
	}

<<<<<<< HEAD
=======

>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
	for _, transaction := range block.Transactions {
		if transaction.Src != 0 &&
				block.AccountBalances[transaction.Src] < transaction.Amount {
			continue
		}
		block.AccountBalances[transaction.Src] -= transaction.Amount
		block.AccountBalances[transaction.Dest] += transaction.Amount
	}
<<<<<<< HEAD
=======
  block.Unlock()
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
}

func blockVerified(blockchain *Blockchain, hash string) {
	blockchain.Lock()
	verifiedBlock := blockchain.unverifiedBlocks[hash]
	delete(blockchain.unverifiedBlocks, hash)
<<<<<<< HEAD
	blockchain.Unlock()
	blockchain.blocks[verifiedBlock.Hash] = verifiedBlock
=======
	blockchain.blocks[verifiedBlock.Hash] = verifiedBlock
  blockchain.Unlock()	
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
	setLongestChain(blockchain, verifiedBlock)
}

func insertTransactionIntoBlock(blockchain *Blockchain, transaction *Transaction) {
<<<<<<< HEAD
	/*
	blockchain.currentBlock.Lock()
	blockchain.currentBlock.Transactions[transaction.TransactionId] = *transaction
	blockchain.currentBlock.Unlock()
	*/
	transactionsBuff := blockchain.transactionsBuff
	transactionsBuff.Lock()
	transactionsBuff.internal[(*transaction).TransactionId] = *transaction
	transactionsBuff.Unlock()
=======
	blockchain.currentBlock.Lock()
	blockchain.currentBlock.Transactions[transaction.TransactionId] = *transaction
	blockchain.currentBlock.Unlock()
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
}

func handleSolved(blockchain *Blockchain, blockHash string, solvedHash string) string {
  return verifyCommand(blockHash, solvedHash)
}

func setLongestChain(blockchain *Blockchain, block *Block) {
<<<<<<< HEAD
	if blockchain.longestChainDepth < block.Depth {
=======
	blockchain.Lock()
  if blockchain.longestChainDepth < block.Depth {
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
		blockchain.longestChainHash = block.Hash
		blockchain.longestChainDepth = block.Depth
		blockchain.currentBlock.PrevHash = block.Hash
	} else if blockchain.longestChainDepth == block.Depth &&
			blockchain.longestChainHash > block.Hash{
		blockchain.longestChainHash = block.Hash
		blockchain.currentBlock.PrevHash = block.Hash
	}
<<<<<<< HEAD
}

func networkTransferBlock(block *Block) string {
	netBlock := &NetworkBlock {
		Transactions:			make([]string, len(block.Transactions)),
		PrevHash:					block.PrevHash,
		Hash:							block.Hash,
		Depth:						block.Depth,
		SolvedHash:				block.SolvedHash,
		AccountBalances:	make([]int, 100),
	}

	for transactionId, _ := range block.Transactions {
		netBlock.Transactions = append(netBlock.Transactions, transactionId)
	}

	for acct, bal := range block.AccountBalances {
		netBlock.AccountBalances[acct] = bal
	}

	jsonBlob, err := json.Marshal(*netBlock)
	if err != nil {
		log.Printf("[-] Block %s could not be seialized into json string\n", netBlock.Hash)
	}
	return fmt.Sprintf("BLOCK_SOLVED %s\n", jsonBlob)
}

func decodeNetworkBlock(message string,
												transactions *ThreadSafeTransactionMap) *Block {
	jsonBlob := strings.TrimSpace(message[12:])
	netBlock := &NetworkBlock{}
	json.Unmarshal([]byte(jsonBlob), &netBlock)

	block := &Block {
		Transactions: make(map[string]Transaction),
		PrevHash: netBlock.PrevHash,
		Hash: netBlock.Hash,
		SolvedHash: netBlock.SolvedHash,
		Depth: netBlock.Depth,
		AccountBalances: make(map[int]int),
	}

	for _, transactionId := range netBlock.Transactions {
		block.Transactions[transactionId] = transactions.internal[transactionId]
	}
	for i, accountBalance := range netBlock.AccountBalances {
		if accountBalance != 0 {
			block.AccountBalances[i] = accountBalance
		}
	}
	return block
=======
  blockchain.Unlock()
>>>>>>> bbde62bbc1a32836300b7f4984f431aec16f507c
}

func blockHash(block *Block) string {
	serializable := &SerialiazableBlock {
		Transactions: block.Transactions,
		PrevHash: block.PrevHash,
		AccountBalances: block.AccountBalances,
	}
	data, err := json.Marshal(serializable)
	if err != nil {
		log.Fatal("[-] Could not parse serialized block into json string")
	}

	str := sha256.Sum256([]byte(data))
	return hex.EncodeToString(str[:])
}

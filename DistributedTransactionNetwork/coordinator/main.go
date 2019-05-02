package main

import (
  "os"
  "sync"
  "fmt"
  "net"
  "log"
  "bufio"
  "strings"
	"strconv"
  "encoding/json"
)

type CoordinatorCommand struct {
  Server string
	Command string
	Key string
  Client string
}

type ThreadSafeHolderMaps struct {
	sync.RWMutex
	keyHolder map[string]string
	readHolder map[string]int
  holderDependant map[string][]string
}

type ThreadSafeClientSet struct {
  sync.RWMutex
  internal map[string]bool
}

func main() {
	fmt.Printf("[+] Starting coordinator\n")
	objects := &ThreadSafeHolderMaps {
		keyHolder: make(map[string]string),
		readHolder: make(map[string]int),
	  holderDependant: make(map[string][]string),
  }
  deadlocks := &ThreadSafeClientSet {
    internal: make(map[string]bool),
  }
  holds := &ThreadSafeClientSet {
    internal: make(map[string]bool),
  }
	go takeCommands(objects, deadlocks, holds)
  listenForServers("4444", objects, deadlocks, holds)
}

func takeCommands(objects *ThreadSafeHolderMaps,
                  deadlocks *ThreadSafeClientSet,
                  holds *ThreadSafeClientSet) {
  reader := bufio.NewReader(os.Stdin)
  for {
    message, err := reader.ReadString('\n')
    if err != nil {
      log.Printf("[+] Error reading from stdin: %v\n", err)
    } else if strings.HasPrefix(message, "KEYHOLDER") {
      objects.RLock()
      for k, v := range objects.keyHolder {
        log.Printf("Key: %s Holder: %s\n", k, v)
      }
      objects.RUnlock()
    } else if strings.HasPrefix(message, "READHOLDER") {
			objects.RLock()
			for k, v := range objects.readHolder {
				log.Printf("%s -> %d\n", k, v)
			}
			objects.RUnlock()
		} else if strings.HasPrefix(message, "DEPS") {
			objects.RLock()
			for k, v := range objects.holderDependant {
				fmt.Printf("%s -> ", k)
				for _, dep := range v {
					fmt.Printf("%s, ", dep)
				}
				fmt.Printf("\n")
			}
			objects.RUnlock()
		}
  }
}

func listenForServers(port string,
                      objects *ThreadSafeHolderMaps,
                      deadlocks *ThreadSafeClientSet,
                      holds *ThreadSafeClientSet) {
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
		go connectionListener(conn, objects, deadlocks, holds)
	}
}

func connectionListener(conn net.Conn, objects *ThreadSafeHolderMaps, deadlocks *ThreadSafeClientSet, holds *ThreadSafeClientSet) {
  reader := bufio.NewReader(conn)
  for {
    message, err := reader.ReadString('\n')
		message = strings.Trim(message, "\n")
		var corrCommand CoordinatorCommand
		jsonErr := json.Unmarshal([]byte(message), &corrCommand)
		if err != nil {
			log.Fatal("[-] Message not read correctly: %v\n", err)
			return
		}
		if jsonErr != nil {
			log.Fatal("[-] Error unmarshalling json: %v\n", err)
		}
    if corrCommand.Command == "SET" {
      deadlocks.RLock()
      _, ok := deadlocks.internal[corrCommand.Client]
      deadlocks.RUnlock()
      if ok {
        deadlocks.Lock()
        conn.Write([]byte("ABORT\n"))
        delete(deadlocks.internal, corrCommand.Client)
        abortNode(corrCommand.Client, objects, holds)
        deadlocks.Unlock()
        continue
      }
      holds.RLock()
      _, ok = holds.internal[corrCommand.Client]
      holds.RUnlock()
      if ok {
        conn.Write([]byte("WAIT\n"))
        continue
      }
      objects.RLock()
      holder, ok := objects.keyHolder[corrCommand.Server + "." + corrCommand.Key]
			readClients, r_ok := objects.readHolder[corrCommand.Server + "." + corrCommand.Key]
      objects.RUnlock()
      if (!ok && (!r_ok || readClients == 0)) || holder == corrCommand.Client {
        //Object is not currently held
        objects.Lock()
        objects.keyHolder[corrCommand.Server + "." + corrCommand.Key] = corrCommand.Client
        objects.Unlock()
        conn.Write([]byte("OK\n"))
      } else {
        conn.Write([]byte("WAIT\n"))
				holds.Lock()
        holds.internal[corrCommand.Client] = true
        holds.Unlock()
        objects.Lock()
				if ok {
					_, ok1 := objects.holderDependant[holder]
					if ok1 {
						appendIfMissing(objects, holder, corrCommand.Client)
					} else {
						objects.holderDependant[holder] = []string{corrCommand.Client}
					}
				}else if r_ok || readClients != 0 {
					for i := 1; i <= 10; i++ {
						bit_offset := uint(i - 1)
						mask := 1 << bit_offset
						if readClients & mask != 0 {
							client := "node" + strconv.Itoa(i)
							_, ok1 := objects.holderDependant[client]
							if ok1 {
								appendIfMissing(objects, client, corrCommand.Client)
							} else {
								objects.holderDependant[client] = []string{corrCommand.Client}
							}
						}
					}
				}
        objects.Unlock()
      }
    } else if corrCommand.Command == "GET" {
      deadlocks.RLock()
      _, ok := deadlocks.internal[corrCommand.Client]
      deadlocks.RUnlock()
      if ok {
        deadlocks.Lock()
        conn.Write([]byte("ABORT\n"))
        delete(deadlocks.internal, corrCommand.Client)
        abortNode(corrCommand.Client, objects, holds)
        deadlocks.Unlock()
        continue
      }
      holds.RLock()
      _, ok = holds.internal[corrCommand.Client]
      holds.RUnlock()
      if ok {
        conn.Write([]byte("WAIT\n"))
        continue
      }
      objects.RLock()
      holder, ok := objects.keyHolder[corrCommand.Server + "." + corrCommand.Key]
      objects.RUnlock()
      if !ok || holder == corrCommand.Client {
        //Object is not currently held
				objects.RLock()
				clientNumber, err := strconv.Atoi(corrCommand.Client[4:])
				if err != nil {
					fmt.Printf("Client name not correctly formatted: %s\n", corrCommand.Client)
				}
				mask := 1 << uint(clientNumber - 1)
				clients, ok := objects.readHolder[corrCommand.Server + "." + corrCommand.Key]
				if !ok {
					objects.readHolder[corrCommand.Server + "." + corrCommand.Key] = 0
					clients = 0
				}
				objects.readHolder[corrCommand.Server + "." + corrCommand.Key] = clients | mask
				objects.RUnlock()

        conn.Write([]byte("OK\n"))
      } else {
        conn.Write([]byte("WAIT\n"))
				holds.Lock()
        holds.internal[corrCommand.Client] = true
        holds.Unlock()
        objects.Lock()
        _, ok1 := objects.holderDependant[holder]
        if ok1 {
          appendIfMissing(objects, holder, corrCommand.Client) 
        } else {
          objects.holderDependant[holder] = []string{corrCommand.Client}
        }
        objects.Unlock()

      }
    } else if corrCommand.Command == "COMMIT" {
      abortNode(corrCommand.Client, objects, holds)  
    } else if corrCommand.Command == "ABORT" {
			abortNode(corrCommand.Client, objects, holds)
		}
  }
}

func traverseDFS(current string, objects *ThreadSafeHolderMaps, deadlocks *ThreadSafeClientSet, visited map[string]bool) bool {
  _, ok := visited[current]
  if ok {
    deadlocks.Lock()
    deadlocks.internal[current] = true
    deadlocks.Unlock()
    return true
  } else {
    visited[current] = true
    deps, ok1 := objects.holderDependant[current]
    if !ok1 {
      return false
    } else {
      for _, elem := range deps {
        if traverseDFS(elem, objects, deadlocks, visited) {
          return true
        }
      }
    }
  }
  return false
}

func appendIfMissing(objects *ThreadSafeHolderMaps, holder string, client string) {
    deps,_ := objects.holderDependant[holder]
    for _, ele := range deps {
        if ele == client {
            return
        }
    }
    objects.holderDependant[holder] = append(deps, client)
}

func abortNode(client string, objects *ThreadSafeHolderMaps, holds *ThreadSafeClientSet) {
  log.Printf("abort client: %s\n", client)
  objects.Lock()
  deps := objects.holderDependant[client]
  holds.Lock()
  for _, dep := range deps {
    delete(holds.internal, dep)
  }
  holds.Unlock()
  delete(objects.holderDependant, client)
  for k, v := range objects.keyHolder {
    if v == client {
      delete(objects.keyHolder, k)
    }
  }
  clientNumber, err := strconv.Atoi(client[4:])
  if err != nil {
    fmt.Printf("Client name not correctly formatted: %s\n", client)
  }
  for k, _ := range objects.readHolder {
    mask := ^(1 << uint(clientNumber - 1))
    objects.readHolder[k] &= mask
  }
  objects.Unlock()
}

func setBit(n int, pos uint) int {
    n |= (1 << pos)
    return n
}

func clearBit(n int, pos uint) int {
    mask := ^(1 << pos)
    n &= mask
    return n
}

func getRemoteAddr(conn net.Conn) string {
	return conn.RemoteAddr().String()
}

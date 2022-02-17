package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var mutex sync.RWMutex
var varTable map[string]string

/*
	Program execution: cd into the directory of server.go,
	Enter command: go run server.go <port #>

	Expected message formats:
		From client:
			get <identifier> \r\n
			set <identifier> <size in bytes> \r\n <byte block> \r\n
		To client:
			Successful set: STORED\r\n
			Failed set: NOT-STORED\r\n
			Get: VALUE <identifier> <number of flags> <size in bytes> \r\n<value>\r\n

*/

func main() {

	//validation on port arg
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: takes one arg (port number)")
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}

	if port < 0 || port >= 65535 {
		fmt.Fprintf(os.Stderr, "Invalid port number)")
		os.Exit(1)
	}
	//end port validation

	//loading varTable from persistent storage if saved previously
	establishVarTable()

	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}

	fmt.Print("** Server initialized, listening on port " + strconv.Itoa(port) + "\n")

	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Print("Connection error with address " + connection.RemoteAddr().String() + "\n")
		}
		go handler(connection)
	}

}

func handler(connection net.Conn) {

	//read connection input, write output back to client
	for {
		b := make([]byte, 512)
		_, err := connection.Read(b)
		if err != nil {
			fmt.Print("Error reading message from address " + connection.RemoteAddr().String() + "\n")
		}

		//processing message
		s := string(b)
		split := strings.Split(s, " ")

		fmt.Print("Received message from " + connection.RemoteAddr().String() + "; Message: " + s + "\n")

		var response string

		if split[0] == "set" {

			tmpSplit := strings.Split(strings.Split(s, "\n")[0], " ")
			size, err := strconv.Atoi(trimString(tmpSplit[len(tmpSplit)-1]))
			if err != nil {
				fmt.Print("Error casting value " + tmpSplit[len(tmpSplit)-1])
			}

			//also truncate val to be the appropriate size in bytes
			tmp := strings.Split(s, "\n")[1]
			var val string
			if size < len(tmp)-1 {
				val = tmp[:size]
			} else {
				val = tmp[:len(tmp)-1]
			}

			//mutex utilized to guarantee map is not read or written at the same time
			mutex.Lock()
			response = setRequest(split[1], val)
			mutex.Unlock()

		} else if split[0] == "get" {

			mutex.RLock()
			response = getRequest(split[1:])
			mutex.RUnlock()

		} else {
			fmt.Print("Unknown command " + split[0] + " from address " + connection.LocalAddr().String() + "\n")
			break
		}

		connection.Write([]byte(response))
	}
	connection.Close()

}

//client is writing a var
func setRequest(id_raw string, val string) string {
	id := trimString(id_raw)
	varTable[id] = trimString(val)

	//write table to persistent storage when mutated
	err := writeVarTable()

	if err != nil {
		return "NOT-STORED\r\n"
	}

	return "STORED\r\n"
}

//client is reading a var
func getRequest(ids []string) string {
	output := ""

	N_FLAGS := 0

	for _, id_raw := range ids {

		id := trimString(id_raw)
		if val, exists := varTable[id]; exists {
			//Note: the flags term was added to work with pymemcache
			output += "VALUE " + id + " " + strconv.Itoa(N_FLAGS) + " " + strconv.Itoa(len(val)) + "\r\n" +
				val + "\r\n"
		}
	}

	return output + "END\r\n"
}

//if map has been written to disk previously, store it in varTable
//else make varTable a new map instance
func establishVarTable() {
	//checking if varTable saved from prior execution
	path, _ := filepath.Abs("../server/varTable.ser")
	_, err := os.Stat(path)

	if err == nil {
		// file exists
		dat, err := os.ReadFile(path)
		b := bytes.NewBuffer(dat)
		d := gob.NewDecoder(b)
		err = d.Decode(&varTable)
		if err != nil {
			panic(err)
		}
	} else {
		// file does not exist; make fresh map
		varTable = make(map[string]string)
	}
}

//writes map to persistent storage
func writeVarTable() error {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	e.Encode(varTable)
	path, _ := filepath.Abs("../server/varTable.ser")
	err := os.WriteFile(path, b.Bytes(), 0644)
	if err != nil {
		return err
	} else {
		return nil
	}
}

//Added this because pymemcache would append a bunch of null characters to key in get requests
//but wouldn't in set requests for some reason
//so this makes them both consistent
func trimString(x string) string {
	x = strings.Trim(x, "\000")
	x = strings.Trim(x, "\n")
	x = strings.Trim(x, "\r")
	return x
}

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

/*
	Program execution: cd into the directory of client.go,
	Enter command: go run client.go <server address>:<server port>

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

	c, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		panic(err)
	}

	for {
		scan := bufio.NewReader(os.Stdin)
		fmt.Printf("Query: ")

		query, _ := scan.ReadString('\n')

		keyword := strings.Split(query, " ")[0]

		if keyword == "exit" {
			fmt.Print("Exiting program... \n")
			c.Close()
			fmt.Print("Connection closed. \n")
			break
		}

		if keyword == "set" {

			fmt.Printf("Value to send: ")
			query1, _ := scan.ReadString('\n')
			query += query1

		}

		fmt.Fprintf(c, trimString(query)+"\r\n")

		var response string

		reader := bufio.NewReader(c)

		switch keyword {
		case "set":
			line, _ := reader.ReadString('\n')
			response = line
			break
		case "get":
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					panic(err)
				}
				response += line
				if strings.Contains(line, "END") {
					break
				}
			}
			break
		}

		fmt.Print(response)

	}
}

//added to trim all messages of trailing characters and guarantee consistency
func trimString(x string) string {
	x = strings.Trim(x, "\000")
	x = strings.Trim(x, "\n")
	x = strings.Trim(x, "\r")
	return x
}

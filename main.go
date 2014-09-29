package main

import (
	"fmt"
	"net"
	"os"
)

func printArgs(args []string) {
	fmt.Println(args)
}

/*
type IP []byte
net.ParseIP(name string) IP
*/
func parseIp(args []string) {
	name := args[0]
	addr := net.ParseIP(name)
	if addr == nil {
		fmt.Println("Invalid address")
	} else {
		fmt.Println("The address is", addr.String())
	}
}

func echoServer(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: echoServer <port>\n")
		os.Exit(1)
	}
	service := ":" + args[0]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
		_, err2 := conn.Write(buf[0:n])
		if err2 != nil {
			return
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

/*
Perform DNS lookups on IP host names
ResolveIPAddr(net, addr string) (*IPAddr, os.Error)
type IPAddr {
  IP IP
}
*/
func ResolveIPAddr(args []string) {
	name := args[0]
	addr, err := net.ResolveIPAddr("ip", name)
	if err != nil {
		fmt.Println("Resolution error", err.Error())
		os.Exit(1)
	}
	fmt.Println("Resolved address is", addr.String())
}

func main() {
	commands := map[string]func(args []string){
		"printArgs":  printArgs,
		"parseIP":    parseIp,
		"resolveIP":  ResolveIPAddr,
		"echoServer": echoServer,
	}

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		os.Exit(1)
	}
	command := os.Args[1]
	fn := commands[command]
	if fn == nil {
		fmt.Fprintf(os.Stderr, "Command '%s' not found\n", command)
		os.Exit(1)
	}
	fn(os.Args[2:])
}

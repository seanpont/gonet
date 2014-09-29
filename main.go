package main

import (
	"fmt"
	"github.com/seanpont/gobro"
	"io/ioutil"
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

/*
Perform DNS lookups on IP host names
ResolveIPAddr(net, addr string) (*IPAddr, os.Error)
type IPAddr {
  IP IP
}
*/
func resolveIp(args []string) {
	name := args[0]
	addr, err := net.ResolveIPAddr("ip", name)
	if err != nil {
		fmt.Println("Resolution error", err.Error())
		os.Exit(1)
	}
	fmt.Println("Resolved address is", addr.String())
}

/*
LookupPort: commonly used ports are listed (on unix machines) in /etc/services.
To interrogate this file, use net.LookupPort(network, service string) (port int, err os.Error)
network = "tcp" or "udp" and the service is the name of the sertice, like "telnet" or "domain"
*/
func lookupPort(args []string) {
	checkArgs(args, 2, "Usage: lookupPort <tcp or udp> <service>")
	networkType, service := args[0], args[1]

	port, err := net.LookupPort(networkType, service)
	gobro.ExitOnError(err)
	fmt.Println("Service port:", port)
}

/*
net.DialTCP(net string, laddr, raddr *TCPAddr) (c *TCPConn, err os.Error)
TCPAddr is a struct containing an IP and a port
type TCPAddr struct {
	IP IP
	Port int
}
TCPConn is a type which allows 'full duplex communication' between client and server.
func (c *TCPConn) Write(b []byte) (n int, err os.Error)
func (c *TCPConn) Read(b []byte) (n int, err os.Error)
This function sends a HEAD request.
*/
func headRequest(args []string) {
	checkArgs(args, 1, "Usage: headRequest host:port")
	service := args[0]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	gobro.ExitOnError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	gobro.ExitOnError(err)

	_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	gobro.ExitOnError(err)

	result, err := ioutil.ReadAll(conn)
	gobro.ExitOnError(err)

	fmt.Println(string(result))
	conn.Close()
}

// ===== ECHO SERVER =========================================================

/*
EchoServer listens to the given port
*/
func echoServer(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: echoServer <port>\n")
		os.Exit(1)
	}
	service := ":" + args[0]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	gobro.ExitOnError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	gobro.ExitOnError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go echoConn(conn)
	}
}

func echoConn(conn net.Conn) {
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

// ===== HELPERS =============================================================

func checkArgs(args []string, numArgs int, message string, a ...interface{}) {
	if len(args) != numArgs {
		fmt.Fprintf(os.Stderr, message+"\n", a...)
		os.Exit(1)
	}
}

// ===== MAIN ================================================================

func main() {
	commands := map[string]func(args []string){
		"printArgs":   printArgs,
		"parseIP":     parseIp,
		"resolveIp":   resolveIp,
		"echoServer":  echoServer,
		"lookupPort":  lookupPort,
		"headRequest": headRequest,
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

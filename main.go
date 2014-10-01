package main

import (
	"bufio"
	"fmt"
	"github.com/seanpont/gobro"
	"io/ioutil"
	"net"
	"os"
	"time"
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

func headRequest2(args []string) {
	checkArgs(args, 1, "Usage: headRequest2 host:port")
	service := args[0]
	conn, err := net.Dial("tcp", service)
	gobro.ExitOnError(err)
	_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	gobro.ExitOnError(err)
	result, err := ioutil.ReadAll(conn)
	gobro.ExitOnError(err)
	fmt.Println(string(result))
}

// ===== UDP DAYTIME =========================================================

func udpDaytimeClient(args []string) {
	checkArgs(args, 1, "udpDaytimeClient host:port")
	service := args[0]
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	gobro.ExitOnError(err)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	gobro.ExitOnError(err)
	_, err = conn.Write([]byte("Time please"))
	gobro.ExitOnError(err)
	var buff [512]byte
	n, err := conn.Read(buff[0:])
	gobro.ExitOnError(err)
	fmt.Println("Time: ", string(buff[:n]))
}

func udpDaytimeServer(args []string) {
	checkArgs(args, 1, "udpDaytimeServer <port>")
	service := ":" + args[0]
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	gobro.ExitOnError(err)
	conn, err := net.ListenUDP("udp", udpAddr)

	gobro.ExitOnError(err)
	for {
		var buf [512]byte
		_, addr, err := conn.ReadFromUDP(buf[0:])
		fmt.Println("readFromUdp", addr)
		if err != nil {
			continue
		}
		go func(conn *net.UDPConn, addr *net.UDPAddr) {
			conn.WriteToUDP([]byte(time.Now().String()), addr)
		}(conn, addr)
	}
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
		go func(conn net.Conn) {
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
		}(conn)
	}
}

func echoServer2(args []string) {
	checkArgs(args, 1, "Usage: echoServer2 <port>")
	listener, err := net.Listen("tcp", ":"+args[0])
	gobro.ExitOnError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go func(conn net.Conn) {
			var buff [512]byte
			n, err := conn.Read(buff[0:])
			if err != nil {
				return
			}
			_, err = conn.Write(buff[:n])
			if err != nil {
				return
			}
		}(conn)
	}
}

func udpEchoServer(args []string) {
	checkArgs(args, 1, "Usage: udpEchoServer <port>")
	packetConn, err := net.ListenPacket("udp", ":"+args[0])
	gobro.ExitOnError(err)
	var b [512]byte
	for {
		n, addr, err := packetConn.ReadFrom(b[0:])
		if err != nil {
			continue
		}
		fmt.Println(string(b[:n]))
		packetConn.WriteTo(b[:n], addr)
	}
}

func udpClient(args []string) {
	checkArgs(args, 1, "udpClient host:port")
	udpAddr, err := net.ResolveUDPAddr("udp", args[0])
	gobro.ExitOnError(err)
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	gobro.ExitOnError(err)
	defer udpConn.Close()
	reader := bufio.NewReader(os.Stdin)
	var buff [512]byte
	for {
		line, _, err := reader.ReadLine()
		gobro.ExitOnError(err)
		_, err = udpConn.Write(line)
		gobro.ExitOnError(err)
		n, err := udpConn.Read(buff[0:])
		gobro.ExitOnError(err)
		fmt.Println("Response: " + string(buff[:n]))
	}
}

// ===== PING ================================================================

// func ping(args []string) {
// 	// THIS FUNCTION IS RETURNING THE FOLLOWING ERROR:
// 	// dial ip4:icmp 206.190.36.45: operation not permitted
// 	checkArgs(args, 1, "ping <host>")
// 	ipAddr, err := net.ResolveIPAddr("ip", args[0])
// 	gobro.ExitOnError(err, "Resolution Error!")
// 	ipConn, err := net.DialIP("ip4:icmp", ipAddr, ipAddr)
// 	gobro.ExitOnError(err)
// 	defer ipConn.Close()
// 	var msg [8]byte
// 	msg[0] = 8  //echo
// 	msg[5] = 13 // identifier
// 	msg[7] = 37 // sequence
// 	check := checkSum(msg[0:8])
// 	msg[2] = byte(check >> 8)
// 	msg[3] = byte(check & 255)

// 	_, err = ipConn.Write(msg[:])
// 	gobro.ExitOnError(err)
// 	ipConn.Read(msg[0:])
// 	fmt.Println(msg)
// }

// func checkSum(msg []byte) uint16 {
// 	sum := 0
// 	for n := 1; n < len(msg)-1; n += 2 {
// 		sum += int(msg[n])*256 + int(msg[n+1])
// 	}
// 	sum = sum>>16 + (sum & 0xffff)
// 	sum += sum >> 16
// 	return uint16(^sum)
// }

func ping(args []string) {
	checkArgs(args, 1, "Usage: ping <host>")

	addr, err := net.ResolveIPAddr("ip", args[0])
	gobro.ExitOnError(err)

	conn, err := net.DialIP("ip4:icmp", addr, addr)
	gobro.ExitOnError(err)

	var msg [512]byte
	msg[0] = 8  // echo
	msg[1] = 0  // code 0
	msg[2] = 0  // checksum, fix later
	msg[3] = 0  // checksum, fix later
	msg[4] = 0  // identifier[0]
	msg[5] = 13 //identifier[1]
	msg[6] = 0  // sequence[0]
	msg[7] = 37 // sequence[1]
	len := 8

	check := checkSum(msg[0:len])
	msg[2] = byte(check >> 8)
	msg[3] = byte(check & 255)

	_, err = conn.Write(msg[0:len])
	gobro.ExitOnError(err)

	_, err = conn.Read(msg[0:])
	gobro.ExitOnError(err)

	fmt.Println("Got response")
	if msg[5] == 13 {
		fmt.Println("identifier matches")
	}
	if msg[7] == 37 {
		fmt.Println("Sequence matches")
	}

	os.Exit(0)
}

func checkSum(msg []byte) uint16 {
	sum := 0

	// assume even for now
	for n := 1; n < len(msg)-1; n += 2 {
		sum += int(msg[n])*256 + int(msg[n+1])
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var answer uint16 = uint16(^sum)
	return answer
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
		"printArgs":        printArgs,
		"parseIP":          parseIp,
		"resolveIp":        resolveIp,
		"echoServer":       echoServer,
		"echoServer2":      echoServer2,
		"lookupPort":       lookupPort,
		"headRequest":      headRequest,
		"headRequest2":     headRequest2,
		"udpDaytimeClient": udpDaytimeClient,
		"udpDaytimeServer": udpDaytimeServer,
		"udpEchoServer":    udpEchoServer,
		"udpClient":        udpClient,
		"ping":             ping,
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

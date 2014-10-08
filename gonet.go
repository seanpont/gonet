package main

import (
	"bufio"
	"code.google.com/p/go.crypto/blowfish"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/seanpont/gobro"
	"io/ioutil"
	"net"
	"os"
	"strings"
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
	gobro.CheckArgs(args, 2, "Usage: lookupPort <tcp or udp> <service>")
	networkType, service := args[0], args[1]

	port, err := net.LookupPort(networkType, service)
	gobro.CheckErr(err)
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
	gobro.CheckArgs(args, 1, "Usage: headRequest host:port")
	service := args[0]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	gobro.CheckErr(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	gobro.CheckErr(err)

	_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	gobro.CheckErr(err)

	result, err := ioutil.ReadAll(conn)
	gobro.CheckErr(err)

	fmt.Println(string(result))
	conn.Close()
}

func headRequest2(args []string) {
	gobro.CheckArgs(args, 1, "Usage: headRequest2 host:port")
	service := args[0]
	conn, err := net.Dial("tcp", service)
	gobro.CheckErr(err)
	_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	gobro.CheckErr(err)
	result, err := ioutil.ReadAll(conn)
	gobro.CheckErr(err)
	fmt.Println(string(result))
}

// ===== UDP DAYTIME =========================================================

func udpDaytimeClient(args []string) {
	gobro.CheckArgs(args, 1, "udpDaytimeClient host:port")
	service := args[0]
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	gobro.CheckErr(err)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	gobro.CheckErr(err)
	_, err = conn.Write([]byte("Time please"))
	gobro.CheckErr(err)
	var buff [512]byte
	n, err := conn.Read(buff[0:])
	gobro.CheckErr(err)
	fmt.Println("Time: ", string(buff[:n]))
}

func udpDaytimeServer(args []string) {
	gobro.CheckArgs(args, 1, "udpDaytimeServer <port>")
	service := ":" + args[0]
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	gobro.CheckErr(err)
	conn, err := net.ListenUDP("udp", udpAddr)

	gobro.CheckErr(err)
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
	gobro.CheckErr(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	gobro.CheckErr(err)

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
	gobro.CheckArgs(args, 1, "Usage: echoServer2 <port>")
	listener, err := net.Listen("tcp", ":"+args[0])
	gobro.CheckErr(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go func(conn net.Conn) {
			var buff [512]byte
			for {
				n, err := conn.Read(buff[0:])
				if err != nil {
					return
				}
				fmt.Printf(string(buff[:]))
				_, err = conn.Write(buff[:n])
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					return
				}
			}
		}(conn)
	}
}

func udpEchoServer(args []string) {
	gobro.CheckArgs(args, 1, "Usage: udpEchoServer <port>")
	packetConn, err := net.ListenPacket("udp", ":"+args[0])
	gobro.CheckErr(err)
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
	gobro.CheckArgs(args, 1, "udpClient host:port")
	udpAddr, err := net.ResolveUDPAddr("udp", args[0])
	gobro.CheckErr(err)
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	gobro.CheckErr(err)
	defer udpConn.Close()
	reader := bufio.NewReader(os.Stdin)
	var buff [512]byte
	for {
		line, _, err := reader.ReadLine()
		gobro.CheckErr(err)
		_, err = udpConn.Write(line)
		gobro.CheckErr(err)
		n, err := udpConn.Read(buff[0:])
		gobro.CheckErr(err)
		fmt.Println("Response: " + string(buff[:n]))
	}
}

// ===== PING ================================================================

func ping(args []string) {
	// THIS FUNCTION IS NOT WORKING
	gobro.CheckArgs(args, 1, "ping <host>")
	ipAddr, err := net.ResolveIPAddr("ip", args[0])
	gobro.CheckErr(err, "Resolution Error!")
	ipConn, err := net.DialIP("ip4:icmp", ipAddr, ipAddr)
	gobro.CheckErr(err)
	defer ipConn.Close()
	var msg [8]byte
	msg[0] = 8  //echo
	msg[5] = 13 // identifier
	msg[7] = 37 // sequence
	check := checkSum(msg[0:8])
	msg[2] = byte(check >> 8)
	msg[3] = byte(check & 255)

	_, err = ipConn.Write(msg[:])
	gobro.CheckErr(err)
	ipConn.Read(msg[0:])
	fmt.Println(msg)
}

func checkSum(msg []byte) uint16 {
	sum := 0
	for n := 1; n < len(msg)-1; n += 2 {
		sum += int(msg[n])*256 + int(msg[n+1])
	}
	sum = sum>>16 + (sum & 0xffff)
	sum += sum >> 16
	return uint16(^sum)
}

// ===== SERIALIZATION =======================================================

type Person struct {
	Name  Name
	Email []Email
}

type Name struct {
	Family   string
	Personal string
}

type Email struct {
	Kind    string
	Address string
}

func defaultPerson() *Person {
	return &Person{
		Name: Name{Family: "Pont", Personal: "Sean"},
		Email: []Email{
			Email{Kind: "work", Address: "sean@cotap.com"},
			Email{Kind: "home", Address: "seanpont@gmail.com"},
		},
	}
}

type Encoder interface {
	Encode(v interface{}) error
}

type Decoder interface {
	Decode(v interface{}) error
}

func serializePerson(args []string) {
	gobro.CheckArgs(args, 1, "serializePersion <json or gob")
	person := defaultPerson()
	fmt.Println(*person)
	file, err := os.Create(os.TempDir() + "temp.json")
	gobro.CheckErr(err)
	var encoder Encoder
	if strings.EqualFold(args[0], "json") {
		encoder = json.NewEncoder(file)
	} else if strings.EqualFold(args[0], "gob") {
		encoder = gob.NewEncoder(file)
	}
	err = encoder.Encode(person)
	gobro.CheckErr(err)
	fmt.Println("Person Persisted")
	file, err = os.Open(file.Name())
	gobro.CheckErr(err)
	var person2 Person
	os.Open(file.Name())
	var decoder Decoder
	if strings.EqualFold(args[0], "json") {
		decoder = json.NewDecoder(file)
	} else if strings.EqualFold(args[0], "gob") {
		decoder = gob.NewDecoder(file)
	}
	err = decoder.Decode(&person2)
	gobro.CheckErr(err)
	fmt.Println(person2)
}

// ===== DIRECTORY BROWSER ===================================================

func ftpServer(args []string) {
	gobro.CheckArgs(args, 1, "ftpServer <port>")
	listener, err := net.Listen("tcp", ":"+args[0])
	gobro.CheckErr(err)
	for {
		conn, err := listener.Accept()
		gobro.CheckErr(err)
		go handleFtpConn(conn)
	}
}

func handleFtpConn(conn net.Conn) {
	defer conn.Close()
	var buff [512]byte
	for {
		n, err := conn.Read(buff[0:])
		if err != nil {
			gobro.LogErr(err)
			return
		}
		request := strings.Split(string(buff[:n]), " ")
		gobro.TrimAll(request)
		fmt.Println(request)
		command := strings.ToLower(request[0])
		fmt.Println(command, command == "cd", command == "ls", command == "pwd")
		var resp string
		if command == "cd" {
			resp, err = "OK", os.Chdir(request[1])
		} else if command == "pwd" {
			resp, err = os.Getwd()
		} else if command == "ls" {
			resp, err = ftpServerLs()
		} else {
			err = errors.New("Unknown command: " + command)
		}
		if err != nil {
			gobro.LogErr(err)
			_, err = conn.Write([]byte(err.Error() + "\r\n\r\n"))
		} else {
			_, err = conn.Write([]byte(resp + "\r\n\r\n"))
		}
		if err != nil { // Write failed
			gobro.LogErr(err)
			return
		}
	}
}

func ftpServerLs() (string, error) {
	dir, err := os.Open(".")
	if err != nil {
		return "", err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return "", err
	}
	return strings.Join(names, "\n"), nil
}

// ===== CRYPTO ==============================================================

func md5Hash(args []string) {
	message := strings.Join(args, " ")
	hash := md5.New()
	_, err := hash.Write([]byte(message))
	gobro.CheckErr(err)
	hashValue := hash.Sum(nil)
	hashSize := hash.Size()
	fmt.Printf("%x\n", hashValue)
	fmt.Println(hashSize)
}

func blowfisher(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: blowfish <pw> <message>")
		os.Exit(1)
	}
	key := args[0]
	message := []byte(strings.Join(args[1:], " "))
	fmt.Printf("Key: %s\n", key)
	fmt.Printf("Message: %s\n", message)

	blockCipher, err := blowfish.NewCipher([]byte(key))
	gobro.CheckErr(err)

	var iv [blowfish.BlockSize]byte
	_, err = rand.Read(iv[:])
	gobro.CheckErr(err)
	fmt.Printf("IV: %x\n", iv)
	stream := cipher.NewCFBEncrypter(blockCipher, iv[:])
	stream.XORKeyStream(message, message)

	fmt.Printf("Encrypted: %x\n", message)

	stream = cipher.NewCFBDecrypter(blockCipher, iv[:])
	stream.XORKeyStream(message, message)

	fmt.Printf("Decrypted: %s\n", string(message))
}

// ===== MAIN ================================================================

func main() {

	gobro.NewCommandMap(
		printArgs,
		parseIp,
		resolveIp,
		echoServer,
		echoServer2,
		lookupPort,
		headRequest,
		headRequest2,
		udpDaytimeClient,
		udpDaytimeServer,
		udpEchoServer,
		udpClient,
		ping,
		serializePerson,
		ftpServer,
		md5Hash,
		blowfisher).Run(os.Args)

}

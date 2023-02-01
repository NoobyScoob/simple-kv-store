package main

// import "sync"
import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	ADDR = "127.0.0.1:59091"
	TYPE = "tcp"
)

type storage struct {
	mu *sync.Mutex
	file *os.File
}

var files = map[string]storage{}
var runWithRandomDelay bool

func main() {
	startServer()
}

func startServer() {
	runWithRandomDelay = len(os.Args) > 1 && (os.Args[1] == "--with-delay")

	// initialize file system map to store data
	initFileSystem()
	closeFileSystemOnExit()

	listener, err := startTcpServer()
	handleErr(err)

	// close server after main returns
	defer listener.Close()
	fmt.Printf("Listening on: %s\n", ADDR)

	handleConnections(listener)
}

// this generates files from a to z
func initFileSystem() {
	fileName := "./storage/0.txt"
	f, _ := os.OpenFile(fileName, os.O_APPEND | os.O_RDWR, 0644)
	files[fileName] = storage{&sync.Mutex{}, f}

	for c := 'a'; c <= 'z'; c++ {
		fileName := "./storage/" + string(c) + ".txt"
		f, _ := os.OpenFile(fileName, os.O_APPEND | os.O_RDWR, 0644)
		files[fileName] = storage{&sync.Mutex{}, f}
	}
}

// spawns go routine to close files on server exit
func closeFileSystemOnExit() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Print("Closing files...")
		files["0.txt"].file.Close()
		for c := 'a'; c <= 'z'; c++ {
			files[string(c) + ".txt"].file.Close()
		}
		os.Exit(0)
	}()
}

// this function starts a tcp server and returns
// listener and error
func startTcpServer() (net.Listener, error) {
	return net.Listen(TYPE, ADDR)
}

// this function handles all the connections
func handleConnections(listener net.Listener) {
	for {
		// accept a connection
		conn, err := listener.Accept()
		handleErr(err)
		// spawning new go routine for each connection
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)

	if err != nil {
		log.Fatal(err)
		conn.Write([]byte(fmt.Sprintf("SERVER_ERROR %v", err)))
		return
	}

	message := string(buffer[:])
	data := strings.Split(message, "\r\n")
	
	command := strings.Split(data[0], " ")
	commandType := strings.ToLower(command[0])

	if runWithRandomDelay {
		rand.Seed(time.Now().UnixNano())
		ms := rand.Intn(500) // random time between 0-500 ms
		fmt.Printf("Server thread sleeping for %d milliseconds.\n", ms)
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}

	if commandType == "gets" || commandType == "get" {
		// do something
		key := command[1]
		value := fetch(key)

		out := "VALUE " + key + " 0 " + fmt.Sprint(len(value)) + "\r\n"
		conn.Write([]byte(out + value + "\r\n" + "END\r\n"))
		
	} else if commandType == "set" {
		key := command[1]
		// numBytes := command[len(command) - 1]
		value := data[1]
		persist(key, value)
		conn.Write([]byte("STORED\r\n"))

	} else {
		// client command not found
		conn.Write([]byte("ERROR\r\n"))
	}

	conn.Close()
}

func persist(key string, value string) {
	fileName := "./storage/" + bucketName(key)
	files[fileName].mu.Lock()
	_, err := files[fileName].file.WriteString(key + " " + value + "\n")
	handleErr(err)
	files[fileName].file.Seek(0, 0)
	files[fileName].mu.Unlock()
}

func fetch(key string) string {
	fileName := "./storage/" + bucketName(key)
	files[fileName].mu.Lock()
	fileScanner := bufio.NewScanner(files[fileName].file)
	value := "keyNotFound"
	// for each line
	for fileScanner.Scan() {
		text := fileScanner.Text()
		data := strings.Split(text, " ")
		if key == data[0] {
			value = data[1]
		}
	}
	files[fileName].file.Seek(0, 0)
	files[fileName].mu.Unlock()
	return value
}

func bucketName(key string) string {
	asciiCode := int((strings.ToLower(key[:1]))[0])
	var fileName string
	if asciiCode >= 97 && asciiCode <= 122 {
		fileName = fmt.Sprintf("%c.txt", asciiCode)
	} else {
		fileName = "0.txt"
	}

	return fileName
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
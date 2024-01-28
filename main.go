package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// FileServer is the file server
type FileServer struct {
	Opts FileServerOptions
}

// FileServerOptions is the options for the file server
type FileServerOptions struct {
	Network string
	Address string
}

// start starts the file server
func (fs *FileServer) start() error {

	// Listen for incoming connections.
	listener, err := net.Listen(fs.Opts.Network, fs.Opts.Address)
	if err != nil {
		return err
	}

	defer listener.Close()

	for {
		// Accept connection on port.
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}

		// Handle connections in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go fs.read(conn)
	}
}

func (fs *FileServer) read(conn net.Conn) {

	// Make a buffer to hold incoming data.
	buf := new(bytes.Buffer)

	// Because we don't know how many bytes we're going to receive, we're going to use CopyN to read the bytes.
	// But we need to know how many bytes to read, we're going to read the size from the connection first.
	// We're going to use binary.Read to read the size from the connection.

	// IMPORTANT: this needs to be declared outside of the loop otherwise
	// it will be redeclared on each iteration of the loop and the loop will never end because the size will always be 0
	var size int64
	binary.Read(conn, binary.LittleEndian, &size)

	for {
		// CopyN copies n bytes (or until an error) from src to dst.
		reqLen, err := io.CopyN(buf, conn, size)
		if err != nil {
			// If we get an error that is EOF, then we reached the end of the file and we can break out of the loop
			if err == io.EOF {
				break
			}

			// If we get an error that is not EOF, then something went wrong and we should exit the function
			log.Fatal(err)
		}

		// Print the read bytes
		fmt.Println(buf.Bytes())

		// Print the number of bytes received
		fmt.Printf("Received %d bytes\n", reqLen)
	}
}

func send(size int) error {

	/*
		NOTE
			This function mimics the client sending a file to the server, however ideally this would be a separate client application.
			The client would connect to the server and send the file. But because this is a single application, we're just going to call the read function directly.
			We are also not going to actually read a file from disk so we're just going to create a file of size and send it to the server.
	*/

	// create a file of size
	file := make([]byte, size)

	// Read the incoming connection into the buffer.
	_, err := io.ReadFull(rand.Reader, file)
	if err != nil {
		return err
	}

	// Connect to the server
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		return err
	}

	/*
		NOTE
			Because we don't know how many bytes we're going to send, we're going to use CopyN to send the bytes.
			But we need to know how many bytes to send, we're going to send the size to the connection first.
			We're going to use binary.Write to write the size to the connection.
	*/
	binary.Write(conn, binary.LittleEndian, int64(size))

	// CopyN copies n bytes (or until an error) from src to dst.
	// It returns the number of bytes copied and the earliest error encountered while copying.
	n, err := io.CopyN(conn, bytes.NewReader(file), int64(size))
	if err != nil {
		return err
	}

	// Print the number of bytes sent
	fmt.Printf("Sent %d bytes\n", n)

	// Close the connection when you're done with it.
	// defer conn.Close()

	return nil
}

func main() {

	go func() {
		// wait 5 seconds before sending the file
		time.Sleep(5 * time.Second)
		send(1024 * 1024 * 10) // 10MB
	}()

	// Create the file server
	server := &FileServer{
		Opts: FileServerOptions{
			Network: "tcp",
			Address: ":3000",
		},
	}

	log.Fatal(server.start())
}

package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/minio/selfupdate"
)

const (
	HOST = "localhost"
	PORT = "80"
	TYPE = "tcp"

	request = `GET / HTTP/1.1
Host: 127.0.0.1
User-Agent: curl/8.9.1
Accept: */*

`
)

func main() {
	tcpServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP(TYPE, nil, tcpServer)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

	defer func() {
		_ = conn.Close()
	}()

	for {
		_, err = conn.Write([]byte(request))
		if err != nil {
			println("Write data failed:", err.Error())
			os.Exit(1)
		}

		// set SetReadDeadline
		err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Println("SetReadDeadline failed:", err)
			// do something else, for example create new conn
			os.Exit(1)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf[:]) // read data from socket
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// time out
				log.Println("read timeout:", err)
			} else {
				// some error else, do something else, for example create new conn
				log.Println("read error:", err)
			}
			os.Exit(1)
		}

		find := false
		for _, k := range bytes.Split(buf, []byte("\n")) {
			if bytes.HasPrefix(k, []byte("Hello from")) {
				fmt.Printf("%s] %s\n", time.Now().Format(time.RFC3339Nano), string(k))
				find = true
				break
			}
		}

		if !find {
			fmt.Printf("%s] Received message [%d]: %s\n", time.Now().Format(time.RFC3339Nano), n, string(buf))
			os.Exit(1)
		}

		time.Sleep(100 * time.Millisecond)

		if err := doUpdate("http://127.0.0.1:8080/"); err != nil {
			log.Println("doUpdate failed:", err)
			os.Exit(1)
		}
	}
}

func main2() {
	tcpServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)

	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP(TYPE, nil, tcpServer)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

	_, err = conn.Write([]byte("This is a message"))
	if err != nil {
		println("Write data failed:", err.Error())
		os.Exit(1)
	}

	// buffer to get data
	received := make([]byte, 1024)
	_, err = conn.Read(received)
	if err != nil {
		println("Read data failed:", err.Error())
		os.Exit(1)
	}

	println("Received message:", string(received))

	conn.Close()
}

func doUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		// error handling
	}

	return err
}

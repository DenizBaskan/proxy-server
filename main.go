package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"proxyserver/socks5"
)

const (
	authenticate = true
	username     = "admin"
	password     = "pass"
	port         = ":8080"
)

func connect(client net.Conn, req RawRequest) {
	/*
		Create a tcp connection between target
		Forward recieved bytes straight to the target
	*/

	target, err := net.Dial("tcp", req.Path) // www.example.com:443
	if err != nil {
		fmt.Printf("Dial fail %v\n", err)
		return
	}
	defer target.Close()

	if _, err := client.Write([]byte(req.Version + " 200 Connection Established\r\n\r\n")); err != nil {
		fmt.Printf("Write fail %v\n", err)
		return
	}

	go io.Copy(client, target)
	io.Copy(target, client)
}

func main() {
	go socks5.Serve()

	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept fail %v\n", err)
			return
		}

		go func() {
			defer func() {
				conn.Close()

				if r := recover(); r != nil {
					fmt.Printf("Panic %v\n", r)
				}
			}()

			reader := bufio.NewReader(conn)
			req_data, err := parse(reader)
			if err != nil {
				fmt.Printf("Parse fail %v\n", err)
				return
			}

			fmt.Printf("%s %s\n", req_data.Method, req_data.Path)

			if authenticate && (req_data.BasicAuth.Username != username || req_data.BasicAuth.Password != password) {
				_, err := conn.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\n" +
					"Proxy-Authenticate: Basic realm=\"Proxy\"\r\n" +
					"Content-Length: 0\r\n" +
					"\r\n"))
				if err != nil {
					fmt.Printf("Write fail %v\n", err)
					return
				}
			}

			if req_data.Method == "CONNECT" {
				connect(conn, req_data)
				return
			}

			req, err := http.NewRequest(req_data.Method, req_data.Path, bytes.NewBuffer(req_data.Body))
			if err != nil {
				fmt.Printf("New request fail %v\n", err)
				return
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("Do request fail %v\n", err)
				return
			}

			res_data := RawResponse{}
			res_data.Headers = make(map[string][]string)

			for k, values := range res.Header {
				for _, value := range values {
					res_data.Headers[k] = append(res_data.Headers[k], value)
				}
			}

			res_data.Body, err = io.ReadAll(res.Body)
			if err != nil {
				fmt.Printf("Read fail %v\n", err)
				return
			}

			res_data.Status = res.Status
			res_data.Version = res.Proto

			conn.Write(build(res_data))
		}()
	}
}



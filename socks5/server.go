package socks5

import (
	"fmt"
	"io"
	"net"
)

const (
	authenticate = true
	username     = "admin"
	password     = "pass"
	port         = ":1080"
)

func Serve() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	for {
		client, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			defer func() {
				client.Close()

				if r := recover(); r != nil {
					fmt.Printf("Panic %v\n", r)
				}
			}()

			buf := make([]byte, 2) // VER, NMETHODS
			if _, err := io.ReadFull(client, buf); err != nil {
				fmt.Printf("Read fail %v\n", err)
				return
			}

			buf = make([]byte, int(buf[1])) // NMETHODS
			if _, err := io.ReadFull(client, buf); err != nil {
				fmt.Printf("Read fail %v\n", err)
				return
			}

			if authenticate {
				_, err = client.Write([]byte{0x05, 0x02}) // authentication username and password
				if err != nil {
					fmt.Printf("Write fail %v\n", err)
					return
				}

				buf = make([]byte, 2) // VER, USERNAME LENGTH
				if _, err := io.ReadFull(client, buf); err != nil {
					fmt.Printf("Read fail %v\n", err)
					return
				}

				username_buf := make([]byte, int(buf[1]))
				if _, err := io.ReadFull(client, username_buf); err != nil {
					fmt.Printf("Read fail %v\n", err)
					return
				}

				buf = make([]byte, 1) // PASSWORD LENGTH
				if _, err := io.ReadFull(client, buf); err != nil {
					fmt.Printf("Read fail %v\n", err)
					return
				}

				password_buf := make([]byte, int(buf[0]))
				if _, err := io.ReadFull(client, password_buf); err != nil {
					fmt.Printf("Read fail %v\n", err)
					return
				}

				if string(username_buf) != username || string(password_buf) != password {
					fmt.Println(string(username_buf), string(password_buf))
					client.Write([]byte{0x01, 0x01}) // STATUS = failure
					return
				}

				client.Write([]byte{0x01, 0x00}) // VER = 1, STATUS = success

			} else {
				_, err = client.Write([]byte{0x05, 0x00}) // no authentication
				if err != nil {
					fmt.Printf("Write fail %v\n", err)
					return
				}
			}

			buf = make([]byte, 4) // VER, CMD, RSV, ATYP
			if _, err := io.ReadFull(client, buf); err != nil {
				fmt.Printf("Read fail %v\n", err)
				return
			}

			atyp := int(buf[3])

			var host string

			switch atyp {
			case 1: // ipv4
				buf = make([]byte, 4)
				_, err = io.ReadFull(client, buf)
				host = net.IP(buf).String()
				break
			case 3: // domain
				buf = make([]byte, 1)
				_, err = io.ReadFull(client, buf)
				buf = make([]byte, int(buf[0]))
				_, err = io.ReadFull(client, buf)
				host = string(buf)
				break
			case 4: // ipv6
				buf = make([]byte, 16)
				_, err = io.ReadFull(client, buf)
				host = net.IP(buf).String()
				break
			}

			if err != nil {
				fmt.Printf("Read fail %v\n", err)
				return
			}

			buf = make([]byte, 2)
			if _, err := io.ReadFull(client, buf); err != nil {
				fmt.Printf("Read fail %v\n", err)
				return
			}

			target_port := int(buf[0])<<8 | int(buf[1])

			fmt.Printf("SOCKS5 %s:%d\n", host, target_port)

			target, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, target_port))
			if err != nil {
				fmt.Printf("Dial fail %v\n", err)
				return
			}

			defer target.Close()

			_, err = client.Write([]byte{
				0x05,       // VER
				0x00,       // REP (Succeeded)
				0x00,       // RSV
				0x01,       // ATYP (IPv4)
				0, 0, 0, 0, // BND.ADDR (IP zero)
				0, 0, // BND.PORT (Port zero)
			})
			if err != nil {
				fmt.Printf("Write fail %v\n", err)
				return
			}

			go io.Copy(client, target)
			io.Copy(target, client)
		}()
	}
}

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

type RawRequest struct {
	Method  string
	Path    string
	Version string
	Headers map[string][]string
	Body    []byte
	BasicAuth struct {
		Username string
		Password string
	}
}

type RawResponse struct {
	Version string
	Status  string
	Headers map[string][]string
	Body    []byte
}

func parse(reader *bufio.Reader) (RawRequest, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return RawRequest{}, err
	}

	var req RawRequest
	req.Headers = make(map[string][]string)

	arr := strings.Split(strings.TrimSuffix(line, "\r\n"), " ")
	req.Method, req.Path, req.Version = arr[0], arr[1], arr[2]

	length := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return RawRequest{}, err
		}

		line = strings.TrimSuffix(line, "\r\n")

		if line == "" {
			break
		}

		arr = strings.SplitN(line, ": ", 2)

		name := strings.ToUpper(arr[0])
		values := append(req.Headers[arr[0]], arr[1])

		if name == "CONTENT-LENGTH" {
			length, _ = strconv.Atoi(arr[1])
		} else if name == "PROXY-AUTHORIZATION" && strings.HasPrefix(arr[1], "Basic ") {
			auth_str, err := base64.StdEncoding.DecodeString(arr[1][6:])
			if err != nil {
				return RawRequest{}, err
			}

			arr := strings.Split(string(auth_str), ":")
			req.BasicAuth.Username, req.BasicAuth.Password = arr[0], arr[1]
		}

		req.Headers[name] = append(values)
	}

	req.Body = make([]byte, length)

	if _, err = reader.Read(req.Body); err != nil {
		return RawRequest{}, err
	}

	return req, nil
}

func build(res RawResponse) []byte {
	s := fmt.Sprintf("%s %s\r\n", res.Version, res.Status)

	for name, values := range res.Headers {
		for _, value := range values {
			s += fmt.Sprintf("%s: %s\r\n", name, value)
		}
	}

	s += "\r\n" + string(res.Body)

	return []byte(s)
}

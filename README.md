# Proxy server
Simple proxy server for HTTP, HTTPS and SOCKS5 that supports user/pass authentication. Curl commands are attatched below to test. Authentication is configurable inside `main.go` (HTTP/HTTPS) and `socks5/server.go` for SOCKS5.

> HTTP

`curl -x http://127.0.0.1:8080 http://www.google.com`

> HTTPS

`curl -x http://127.0.0.1:8080 https://www.google.com`

> SOCKS5

`curl --proxy socks5h://admin:pass@127.0.0.1:1080 https://www.google.com`

### Run
`go run .`

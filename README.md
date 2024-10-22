## gdos
gdos is a simple HTTP Denial of Service tool written in Golang. It is designed to be simple and easy to use. It is not meant to be used for malicious purposes. It is meant to be used for testing the performance of a web server. It is designed to just affect threaded servers.

### Whats going on inside gdos?

1. gdos creates a number of threads that will send a request to the server.
2. it periodically send some random data to the server through the opened connections.
3. it will try to keep the connections open as long as possible. if the server closes the connection, gdos will try to open a new connection.

### How to use gdos?

1. Clone the repository
```
git clone https://github.com/ostadgeorge/gdos.git && cd gdos
```
2. Run `go build`
3. Run `./gdos -host <HOST> -port <PORT> -numSockets <NUM_SOCKETS>`. check code for more options.
4. To stop the attack, press `Ctrl + C`

### Example
1. Run simple HTTP server on port 8000
```
python -m http.server 8080
```
2. Run gdos
```
./gdos -host 127.0.0.1 -port 8080 -numSockets 20000
```
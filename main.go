package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

const (
	SAMPLES_SIZE       = 10000
	NODES_FILE         = "nodes.txt"
	MAX_DATAGRAMA_SIZE = 1024
)

var (
	nodes []string
)

func FindShortestPath(ori string, dest string, conn *net.UDPConn) (*ResponsePayload, *time.Duration, error) {
	requestPayload := RequestPayload{
		Ori:  ori,
		Dest: dest,
	}
	jsonRequest, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, nil, err
	}
	start := time.Now()
	if _, err := conn.Write(jsonRequest); err != nil {
		return nil, nil, err
	}

	jsonResponse := make([]byte, MAX_DATAGRAMA_SIZE)
	n, _, err := conn.ReadFromUDP(jsonResponse)
	if err != nil {
		return nil, nil, err
	}
	jsonResponse = jsonResponse[:n]
	responsePayload := ResponsePayload{}
	if err := json.Unmarshal(jsonResponse, &responsePayload); err != nil {
		return nil, nil, err
	}
	rtt := time.Since(start) - responsePayload.CalcDuration
	return &responsePayload, &rtt, nil
}

func main() {
	host, ok := os.LookupEnv("HOST")
	if !ok {
		log.Fatal("undefined PORT")
	}

	if err := setup(); err != nil {
		log.Fatal(err)
	}

	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer closeUDPConnection(conn)

	var samples []time.Duration
	for i := 0; i < SAMPLES_SIZE; i++ {
		rand.Seed(time.Now().UnixNano())
		ori := nodes[rand.Intn(len(nodes))]
		dest := nodes[rand.Intn(len(nodes))]

		log.Printf("sending request to find the shortest path between %s and %s", ori, dest)
		res, rtt, err := FindShortestPath(ori, dest, conn)
		if err != nil {
			log.Fatal(err)
		}
		samples = append(samples, *rtt)
		log.Printf("shortest path received %v", res.Path)
	}

	var mean float64
	for _, sample := range samples {
		mean += float64(sample)
	}
	mean = mean / float64(len(samples))

	var sd float64
	for _, sample := range samples {
		sd += math.Pow((float64(sample) - mean), 2)
	}
	sd = math.Sqrt(sd / float64(len(samples)))

	log.Printf("average RTT is %.2f (+- %.2f)", mean, sd)
}

func setup() error {
	file, err := os.ReadFile(NODES_FILE)
	if err != nil {
		return err
	}

	nodes = strings.Split(string(file), " ")

	return nil
}

type RequestPayload struct {
	Ori  string `json:"ori"`
	Dest string `json:"dest"`
}

type ResponsePayload struct {
	Path         []string      `json:"path"`
	CalcDuration time.Duration `json:"calc-duration"`
}

type ResponseErrorPayload struct {
	Message string `json:"message"`
}

func closeUDPConnection(conn *net.UDPConn) {
	err := conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

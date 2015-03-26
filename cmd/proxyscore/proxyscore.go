package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

const servicePort = 4673

func main() {
	proxy := flag.String("p", "", "valid proxy ip address with port (ex: 1.2.3.4:6789)")
	service := flag.String("r", "", "receiver (with service running) ip address ")

	flag.Parse()

	ip, proxyport, errsplit := net.SplitHostPort(*proxy)
	if errsplit != nil {
		log.Fatalf("cannot split proxy address into ip and port: %s", errsplit)
	}

	proxyip := net.ParseIP(ip)
	serviceip := net.ParseIP(*service)

	if proxyip == nil || serviceip == nil {
		log.Fatalf("invalid ip address provided")
	}

	err := os.Setenv("HTTP_PROXY", net.JoinHostPort(proxyip.String(), proxyport))
	if err != nil {
		log.Fatalf("Error setting proxy in env: %s", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%d", serviceip, servicePort))
	defer resp.Body.Close()

	json.NewEncoder(os.Stdout).Encode(resp.Request)

	if err != nil {
		log.Fatalf("Error contacting service: %s", err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(body)
}

package main

import (
	"bytes"
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

var serviceIp = []byte{0x34, 0x35, 0x2e, 0x35, 0x35, 0x2e, 0x31, 0x34, 0x34, 0x2e, 0x31, 0x36, 0x31}

func main() {
	proxy := flag.String("p", "", "valid proxy ip address with port (ex: 1.2.3.4:6789)")
	verbose := flag.Bool("v", false, "verbose in case of errors")

	flag.Parse()

	ip, proxyport, errsplit := net.SplitHostPort(*proxy)
	if errsplit != nil {
		log.Fatalf("cannot split proxy address into ip and port: %s", errsplit)
	}

	proxyip := net.ParseIP(ip)

	if proxyip == nil {
		log.Fatalf("invalid ip address provided")
	}

	err := os.Setenv("HTTP_PROXY", net.JoinHostPort(proxyip.String(), proxyport))
	if err != nil {
		log.Fatalf("Error setting proxy in env: %s", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%d", serviceIp, servicePort))
	if err != nil {
		log.Fatalf("Error contacting service: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if status := resp.StatusCode; status < 200 || status > 299 {
		log.Printf("HTTP NOK - Received %d\n", status)
		if *verbose {
			log.Printf("BODY: %s\n", bytes.NewBuffer(body).String())
		}
		os.Exit(1)
	}

	var out bytes.Buffer
	json.Indent(&out, body, "", "\t")
	out.WriteTo(os.Stdout)
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
)

const servicePort = 4673

var serviceIp = []byte{0x34, 0x35, 0x2e, 0x35, 0x35, 0x2e, 0x31, 0x34, 0x34, 0x2e, 0x31, 0x36, 0x31}

type serviceResponse struct {
	Header     map[string][]string
	RemoteAddr string
}

func (s serviceResponse) String() string {
	return fmt.Sprintf("%#v", s)
}

func main() {
	proxy := flag.String("p", "", "valid proxy ip address with port (ex: 1.2.3.4:6789)")
	verbose := flag.Bool("v", false, "verbose in case of errors")

	flag.Parse()

	resp, err := http.Get("http://api.ipify.org/?format=json")
	if err != nil {
		log.Fatalf("Error in service to get your IP: %s", err)
	}
	defer resp.Body.Close()

	d := struct{ IP string }{}
	json.NewDecoder(resp.Body).Decode(&d)
	myip := net.ParseIP(d.IP)
	if myip == nil {
		log.Fatalf("Error parsing your ip %s", d.IP)
	}

	ip, proxyport, errsplit := net.SplitHostPort(*proxy)
	if errsplit != nil {
		log.Fatalf("cannot split proxy address into ip and port: %s", errsplit)
	}

	proxyip := net.ParseIP(ip)

	if proxyip == nil {
		log.Fatalf("invalid ip address provided")
	}

	proxyurl, errurl := url.Parse(fmt.Sprintf("http://%s", net.JoinHostPort(proxyip.String(), proxyport)))
	if errurl != nil {
		log.Fatalf("Cannot construct proxy url %s", errurl)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyurl),
	}
	client := &http.Client{Transport: transport}

	resp, errget := client.Get(fmt.Sprintf("http://%s:%d", serviceIp, servicePort))
	if errget != nil {
		log.Fatalf("Error contacting service: %s", errget)
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status < 200 || status > 299 {
		log.Fatalf("HTTP NOK - Received %d\n", status)
	}

	result := serviceResponse{}
	errdec := json.NewDecoder(resp.Body).Decode(&result)
	if errdec != nil {
		log.Fatalf("Error decoding incoming json: %s", errdec)
	}
	if *verbose {
		log.Println(result)
	}

	guilty := containsMyIP(result, myip)

	if len(guilty) == 0 {
		fmt.Printf("\n### Your IP %s was not detected\n", myip)
	} else {
		fmt.Printf("\n### Your IP %s was detected on the target server as %s\n", myip, guilty)
	}

}

func containsMyIP(result serviceResponse, myip net.IP) []string {
	var guilty []string
	for k, v := range result.Header {
		for _, vv := range v {
			if ip := net.ParseIP(vv); ip != nil && myip.Equal(ip) {
				guilty = append(guilty, k)
			}
		}
	}

	remoteip, _, _ := net.SplitHostPort(result.RemoteAddr)
	if ip := net.ParseIP(remoteip); ip != nil && myip.Equal(ip) {
		guilty = append(guilty, "RemoteAddr")
	}

	return guilty
}

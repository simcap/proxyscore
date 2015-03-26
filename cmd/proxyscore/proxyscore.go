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

var serviceURL = "\x68\x74\x74\x70\x3a\x2f\x2f\x34\x35\x2e\x35\x35\x2e\x31\x34\x34\x2e\x31\x36\x31\x3a\x34\x36\x37\x33"
var myIpURL = "http://api.ipify.org/?format=json"

type serviceResponse struct {
	Header     map[string][]string
	RemoteAddr string
}

func (s serviceResponse) String() string {
	return fmt.Sprintf("%#v", s)
}

func main() {
	proxy := flag.String("p", "", "valid proxy ip address with port (ex: 1.2.3.4:6789)")
	verbose := flag.Bool("v", false, "verbose mode")

	flag.Parse()

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

	resp, errget := client.Get(serviceURL)
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

	myip := getMyIp()
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

func getMyIp() net.IP {
	resp, err := http.Get(myIpURL)
	if err != nil {
		log.Fatalf("Error '%s' service to get your IP: %s", myIpURL, err)
	}
	defer resp.Body.Close()

	d := struct{ IP string }{}
	json.NewDecoder(resp.Body).Decode(&d)
	myip := net.ParseIP(d.IP)
	if myip == nil {
		log.Fatalf("Error parsing your ip %s from service %s", d.IP, myIpURL)
	}
	return myip
}

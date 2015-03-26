package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

func main() {
	proxy := flag.String("p", "", "valid proxy ip address with port (ex: 1.2.3.4:6789)")
	flag.Parse()

	client := NewProxiedClient(*proxy)

	resp := client.pingTarget(targetURL)
	defer resp.Body.Close()

	received := Received{}
	errdec := json.NewDecoder(resp.Body).Decode(&received)
	if errdec != nil {
		log.Fatalf("Error decoding incoming json: %s", errdec)
	}

	myip := MyIp()
	ipdetected := received.containsIP(myip)
	proxyinfodetected := received.containsProxyInfo()

	outcome := NewOutcome(myip, *proxy, ipdetected, proxyinfodetected)

	j, _ := json.MarshalIndent(outcome, "", "  ")
	os.Stdout.Write(j)
}

type Outcome struct {
	Anonymous      bool
	Score          int
	MyIP           string
	Proxy          string
	IPdetection    map[string]string
	Proxydetection map[string]string
}

func NewOutcome(ip net.IP, proxy string, ipdetection map[string]string, proxydetection map[string]string) Outcome {
	anon := false
	if ipdetection == nil || len(ipdetection) == 0 {
		anon = true
	}
	if proxydetection == nil || len(proxydetection) == 0 {
	}
	return Outcome{Anonymous: anon, Proxy: proxy, MyIP: ip.String(), IPdetection: ipdetection, Proxydetection: proxydetection}
}

type ProxiedClient struct {
	*http.Client
}

func NewProxiedClient(proxy string) *ProxiedClient {
	ip, _, err := net.SplitHostPort(proxy)
	if err != nil {
		log.Fatalf("cannot split proxy address into ip and port: %s", err)
	}

	if proxyip := net.ParseIP(ip); proxyip == nil {
		log.Fatalf("invalid ip address provided")
	}

	proxyurl, errurl := url.Parse("http://" + proxy)
	if errurl != nil {
		log.Fatalf("Cannot parse proxy url %s", errurl)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyurl),
	}
	return &ProxiedClient{&http.Client{Transport: transport}}
}

func (pc *ProxiedClient) pingTarget(url string) *http.Response {
	resp, errget := pc.Get(url)
	if errget != nil {
		log.Fatalf("Error contacting target at %s: %s", url, errget)
	}

	if status := resp.StatusCode; status < 200 || status > 299 {
		log.Fatalf("HTTP NOK - Received %d\n", status)
	}

	return resp
}

type Received struct {
	Headers    map[string][]string
	RemoteAddr string
}

func (r *Received) containsIP(myip net.IP) map[string]string {
	var detected map[string]string
	for k, v := range r.Headers {
		for _, vv := range v {
			if ip := net.ParseIP(vv); ip != nil && myip.Equal(ip) {
				detected[k] = ip.String()
			}
		}
	}

	remoteip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip := net.ParseIP(remoteip); ip != nil && myip.Equal(ip) {
		detected["RemoteAddr"] = ip.String()
	}

	return detected
}

var proxyinforeg = regexp.MustCompile(`(?i)forw|via|prox`)

func (r *Received) containsProxyInfo() map[string]string {
	var detected map[string]string
	for k, v := range r.Headers {
		if proxyinforeg.MatchString(k) {
			detected[k] = v[0]
		}
	}

	return detected
}

func MyIp() net.IP {
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

var targetURL = "\x68\x74\x74\x70\x3a\x2f\x2f\x34\x35\x2e\x35\x35\x2e\x31\x34\x34\x2e\x31\x36\x31\x3a\x34\x36\x37\x33"
var myIpURL = "http://api.ipify.org/?format=json"

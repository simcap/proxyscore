package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var (
	proxyFlag             = flag.String("p", "", "valid proxy ip address with port (ex: 176.107.17.129:8080")
	receiverTargetURLFlag = flag.String("t", "", "Target URL that has the receiver deployed to analyzes the request and send back results")
)

func main() {
	flag.Parse()

	client := NewProxiedClient(*proxyFlag)

	resp := client.pingTarget(*receiverTargetURLFlag)
	defer resp.Body.Close()

	received := Received{}
	errdec := json.NewDecoder(resp.Body).Decode(&received)

	if errdec != nil {
		log.Fatalf("Error decoding incoming json: %s", errdec)
	}

	myip := myIP()
	ipdetected := received.containsIP(myip)
	proxyinfodetected := received.containsProxyInfo()

	outcome := NewOutcome(myip, *proxyFlag, ipdetected, proxyinfodetected)

	j, _ := json.MarshalIndent(outcome, "", "  ")
	os.Stdout.Write(j)
}

type Outcome struct {
	Score          int
	Level          string
	MyIP           string
	Proxy          string
	IPdetection    map[string]string
	Proxydetection map[string]string
}

func NewOutcome(ip net.IP, proxy string, ipdetection map[string]string, proxydetection map[string]string) Outcome {
	ipdetected, proxydetected := true, true
	score, level := 3, "transparent"

	if ipdetection == nil || len(ipdetection) == 0 {
		ipdetected = false
	}
	if proxydetection == nil || len(proxydetection) == 0 {
		proxydetected = false
	}

	if !ipdetected {
		if proxydetected {
			level, score = "anonymous", 2
		} else {
			level, score = "elite", 1
		}
	}

	return Outcome{
		Proxy: proxy, Level: level,
		Score: score,
		MyIP:  ip.String(), IPdetection: ipdetection,
		Proxydetection: proxydetection,
	}
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
	Header     map[string][]string
	RemoteAddr string
}

func (r *Received) containsIP(myip net.IP) map[string]string {
	detected := make(map[string]string)
	for k, v := range r.Header {
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

var proxyinforeg = regexp.MustCompile(`(?i)forw|via|prox|client|ip`)

func (r *Received) containsProxyInfo() map[string]string {
	detected := make(map[string]string)
	for k, v := range r.Header {
		if proxyinforeg.MatchString(k) {
			detected[k] = v[0]
		}
	}

	return detected
}

func myIP() net.IP {
	service := "http://checkip.amazonaws.com/"
	resp, err := http.Get(service)
	if err != nil {
		log.Fatalf("cannot get your ip from service %s: %s", service, err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("cannot parse your ip '%q' from service %s: %s", b, service, err)
	}
	defer resp.Body.Close()

	return net.ParseIP(strings.TrimSpace(string(b)))
}

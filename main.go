package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var IP_PROVIDER = "https://ip.home.wmtech.cc/"

type Record struct {
	Data     string `json:"data"`
	Port     int64  `json:"port"`
	Priority int64  `json:"priority"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	TTL      int64  `json:"ttl"`
	Weight   int64  `json:"weight"`
}

func getOwnIPv6() (string, error) {
	resp, err := http.Get(IP_PROVIDER)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return buf.String(), nil
}

func getDomainIPv6() (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/AAAA/%s", DOMAIN, SUBDOMAIN), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", GODADDY_KEY, GODADDY_SECRET))
	c := new(http.Client)
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	in := make([]struct {
		Data string `json:"data"`
	}, 1)
	json.NewDecoder(resp.Body).Decode(&in)
	if len(in) == 0 {
		return "", nil
	}
	return in[0].Data, nil
}

func putNewIP(ip string) error {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode([]Record{{
		strings.TrimSuffix(ip, "\n"),
		65535,
		0,
		"",
		"",
		600,
		0,
	}})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT",
		fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/AAAA/%s", DOMAIN, SUBDOMAIN),
		&buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", GODADDY_KEY, GODADDY_SECRET))
	c := new(http.Client)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		return nil
	} else {
		return fmt.Errorf("failed with HTTP status code %d", resp.StatusCode)
	}
}

func run() {
	ownIP, err := getOwnIPv6()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ownIP)
	domainIP, err := getDomainIPv6()
	if err != nil {
		log.Fatal(err)
	}
	if domainIP == "" || domainIP != ownIP {
		if err := putNewIP(ownIP); err != nil {
			log.Fatal(err)
		}
	}
}

// globals
var GODADDY_KEY = ""
var GODADDY_SECRET = ""
var DOMAIN = ""
var SUBDOMAIN = "test"

func main() {
	// log file flag
	logFile := "godaddns.log"
	POLLING := 360 // Polling interval in seconds. Lookup Godaddy's current rate limits before setting too low
	DOMAIN = "beautifood.io"
	GODADDY_SECRET = os.Getenv("API_SECRET")
	GODADDY_KEY = os.Getenv("API_KEY")
	SUBDOMAIN = os.Getenv("SUBDOMAIN")

	if logFile == "" {
		log.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Couldn't open log file: %s", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	if DOMAIN == "" {
		log.Fatalf("You need to provide your domain")
	}

	if GODADDY_SECRET == "" {
		log.Fatalf("You need to provide your API secret")
	}

	if GODADDY_KEY == "" {
		log.Fatalf("You need to provide your API key")
	}

	// run
	for {
		run()
		time.Sleep(time.Second * time.Duration(POLLING))
	}
}

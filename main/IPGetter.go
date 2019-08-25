package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
    "crypto/tls"
)

// url for
const (
	IPV4 = "https://ipv4.lookup.test-ipv6.com/ip/"
	IPV6 = "https://ipv6.lookup.test-ipv6.com/ip/"
)

// IPInfo {"ip":"76.68.55.226","type":"ipv4","subtype":"",
// "via":"","padding":"","asn":"577","asnlist":"577",
// "asn_name":"Bell Canada","country":"CA","protocol":"HTTP/2.0"}
type IPInfo struct {
	IP string  `json:"ip"`
	Type string `json:"type"`
	Subtype string `json:"subtype"`
	Via string `json:"via"`
	Padding string `json:"padding"`
	Asn string `json:"asn"`
	Asnlist string `json:"asnlist"`
	Country string `json:"country"`
	Protocal string `json:"protocal"`
}

var client *http.Client

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
}

// GetIPv4 get ipv4
func GetIPv4() string {
	resp, err := client.Get(IPV4)
	if err != nil {
		Logger.Print(err.Error())
		return ""
	}
	Debug.Printf("test IPv4: %#v\n", resp)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	Debug.Printf("test IPv4 body: %s\n", body)

	res := IPInfo{}
	json.Unmarshal([]byte(body), &res)

	Debug.Printf("IPv4: %s\n", res.IP)
	return res.IP
}

// GetIPv6 get ipv6
func GetIPv6() string {
	resp, err := client.Get(IPV6)
	if err != nil {
		Logger.Print(err.Error())
		return ""
	}
	Debug.Printf("test IPv6: %#v\n", resp)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	Debug.Printf("test IPv6 body: %s\n", body)

	res := IPInfo{}
	json.Unmarshal([]byte(body), &res)
	
	Debug.Printf("IPv6: %s\n", res.IP)
	return res.IP
}
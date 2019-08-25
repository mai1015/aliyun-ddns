package main

import (
	"flag"
    "log"
	"os"
	"time"
	"strings"
	"strconv"
	"io/ioutil"
)

// Logger default logger
var (
	Debug = log.New(ioutil.Discard, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	Logger = log.New(os.Stdout, "[MAIN] ", log.Ldate|log.Ltime|log.Lshortfile)
)

var (
	t int

	interval int
	ts []string
	rr []string
	domain string
	ttl int
	delete bool
	value string
)

var (
	ipv4 string
	ipv6 string
)

func init() {
	if os.Getenv("DNS_ENV") != "PRODUCTION" {
		Debug.SetOutput(os.Stdout)
	}
	// flag
	var region, u, p, r, types string
	flag.StringVar(&region, "e", getenv("REGION", "default"), "region of service.")
	flag.StringVar(&u, "u", os.Getenv("AKID"), "Aliyun ID")
	flag.StringVar(&p, "p", os.Getenv("AKSCT"), "Aliyun secret")
	flag.StringVar(&domain, "d", os.Getenv("DOMAIN"), "root domain name")
	flag.StringVar(&r, "r", os.Getenv("RR"), "list of sub-domain name. Can split with \",\".")
	flag.StringVar(&types, "t", getenv("TYPE", "A"), "list of type need to update. Can split with \",\".")
	flag.IntVar(&interval, "i", getenvInt("INTERVAL", -1), "time interval. -1 means it will only run once.")
	flag.IntVar(&ttl, "l", getenvInt("TTL", 600), "time")
	flag.BoolVar(&delete, "x", false, "delete domains from rr")
	flag.StringVar(&value, "v", "", "set value")
	flag.Parse()
	// dealing

	InitDomain(region, u, p)
	if (len(r) == 0 || len(domain) == 0) {
		Logger.Fatalf("Value domain, rr should not be empty\n")
		os.Exit(1)
	}

	rr = strings.Split(r, ",")
	for i, or := range rr {
		rr[i] = strings.TrimSpace(or)
	}
	ts = strings.Split(types, ",")
	for i, ot := range ts {
		ts[i] = strings.TrimSpace(ot)
	}
}

func getenvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
    }
    return fallback
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func main() {
	if delete {
		Logger.Println("Staring to delete")
		for _, r := range rr {
			for _, t := range ts {
				resp, _ := DelDomainRecord(domain, r, t)
				if resp != nil {
					Logger.Printf("remove %s record(s) from domain %s for %s type.\n", resp.TotalCount, resp.RR + "." + domain, t)
				} else {
					Logger.Printf("failed to remove record(s) from domain %s.\n", r + "." + domain)
				}
			}
		}
		return
	}

	if len(value) != 0 {
		Logger.Println("Staring to manual set")
		for _, r := range rr {
			for _, t := range ts {
				resp, _ := AddDomainRecord(domain, r, ttl, t, value)
				if resp != nil {
					Logger.Printf("Successfully add domain %s record %s for %s at default line.\n", t, value,  r + "." + domain)
				} else {
					Logger.Printf("failed to add record(s) from domain %s.\n", r + "." + domain)
				}
			}
		}
		return
	}

	Logger.Println("Staring")
	for true {
		routing()
		if interval == -1 {
			break
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func routing() {
	Logger.Printf("Start task\n")
	ipv4 = GetIPv4()
	Logger.Printf("current ipv4: %s\n", ipv4)
	ipv6 = GetIPv6()
	Logger.Printf("current ipv6: %s\n", ipv6)
	for _, r := range rr {
		for _, t := range ts {
			doDomain(r, t)
		}
	}
	t = t + 1
}

func doDomain(r, t string) {
	Debug.Printf("doDomain(%s, %s)", r, t)
	ip := getIP(t)
	if len(ip) <= 0 {
		Logger.Printf("no valid ip address! cant set %s type for %s.\n", t, r + "." + domain)
		return
	}
	Debug.Printf("try to update domain %s for %s type.\n", r + "." + domain, t)
	resp, err := GetDomainRecord(domain, r, t)
	if err != nil {
		Logger.Printf("%s: %s\n", r + "." + domain, err.Error())
		return
	}

	if (resp.IsSuccess()) {
		if resp.TotalCount <= 0 {
			// add record
			Logger.Printf("found 0 record for %s in default line, try to add record.\n", r + "." + domain)
			response, _ := AddDomainRecord(domain, r, ttl, t, ip)
			if response != nil && response.IsSuccess() {
				Logger.Printf("Successfully add domain record for %s at default line.\n", r + "." + domain)
			} else {
				Logger.Printf("Failed to add record for %s.\n", r + "." + domain)
			}
		} else {
			Debug.Printf("found existing %s", r + "." + domain)
			for _, record := range resp.DomainRecords.Record {
				if record.RR == r && record.DomainName == domain && 
				   record.Type == t && record.Line == "default" {
					if record.Value == ip {
						Logger.Printf("found same ip record for %s at default line for id %s.\n", r + "." + domain, record.RecordId)
						break
					}
					Logger.Printf("try to update domain %s: %s", r + domain, record.RecordId)
					response, _ := UpdateDomainRecord(record.RecordId, r, ttl, t, ip)
					if response != nil && response.IsSuccess() {
						Logger.Printf("Successfully update domain record for %s at default line for id %s.\n", r + "." + domain, response.RecordId)
					} else {
						Logger.Printf("Failed to update record for %s.\n", r + "." + domain)
					}
					break
				}
			}
		}
	} else {
		Logger.Printf("Request failed for: %s", r + "." + domain)
	}
}

func getIP(recordType string) string {
	switch recordType {
	case "AAAA":
		return ipv6
	default:
		return ipv4
	}
}
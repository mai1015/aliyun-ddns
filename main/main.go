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
	get bool
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
	flag.BoolVar(&get, "g", false, "get values")
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
	if delete || get || len(value) != 0{
		Logger.Println("Staring to manual task")
		for _, r := range rr {
			for _, t := range ts {
				if delete {
					DelDomainRecord(domain, r, t)
				} else if get {
					resp, _ := GetDomainRecord(domain, r, t)
					Logger.Println(resp.DomainRecords)
				} else if len(value) != 0{
					AddDomainRecord(domain, r, ttl, t, value)
				}
			}
		}
		return
	}

	Logger.Println("Staring task")
	for true {
		Logger.Println("Perform update")
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
	if err != nil {return}

	if (resp.IsSuccess()) {
		if resp.TotalCount <= 0 {
			// add record
			Logger.Printf("found 0 record for %s in default line, try to add record.\n", r + "." + domain)
			AddDomainRecord(domain, r, ttl, t, ip)
		} else {
			Debug.Printf("found existing %s", r + "." + domain)
			for _, record := range resp.DomainRecords.Record {
				if record.RR == r && record.DomainName == domain && 
				   record.Type == t && record.Line == "default" {
					if record.Value == ip {
						Logger.Printf("found same ip record for %s at default line for id %s.\n", r + "." + domain, record.RecordId)
						break
					}
					Debug.Printf("try to update domain %s: %s", r + "." + domain, record.RecordId)
					UpdateDomainRecord(record.RecordId, r, ttl, t, ip)
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
	case "A":
		return ipv4
	default:
		return ""
	}
}
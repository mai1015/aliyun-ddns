package main

import (
	"os"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
)

const (
	line = "default"
)

// dns client
var (
	DDNSClient *alidns.Client
)

// type ClientWapper struct {
// 	client *alidns.Client
// }

// InitDomain for region
func InitDomain(region, key, secret string) {
	Debug.Printf("InitDomain(%s, %s, %s)\n", region, key, secret)

	if (len(key) == 0 || len(secret) == 0) {
		os.Exit(1)
	}

	var err error
	DDNSClient, err = alidns.NewClientWithAccessKey(region, key, secret)
	if err != nil {
		Logger.Fatalln(err.Error())
		os.Exit(1)
	}
}

// GetDomainRecord get one record
func GetDomainRecord(domain string, rr string, recordType string)  (response *alidns.DescribeSubDomainRecordsResponse, err error) {
	Debug.Printf("getRecord(%s, %s, %s)\n", domain, rr, recordType)
	request := alidns.CreateDescribeSubDomainRecordsRequest()

	request.Scheme = "https"
	request.SubDomain = rr + "." + domain
	request.Type = recordType
	request.Line = line
	request.PageSize = requests.NewInteger(10)

	response, err = DDNSClient.DescribeSubDomainRecords(request)
	if err != nil {
		Logger.Printf("Failed to get domain %s for %s type: %s\n", rr + "." + domain, recordType, err.Error())
	} else {
		Logger.Printf("Successfully got domain records from %s for %s type: %d\n", rr + "." + domain, recordType, response.TotalCount)
	}
	Debug.Printf("response is %#v\n", response)
	return response, err
}

// DelDomainRecord delete
func DelDomainRecord(domain string, rr string, recordType string) (response *alidns.DeleteSubDomainRecordsResponse, err error) {
	Debug.Printf("delRecord(%s, %s, %s)\n", domain, rr, recordType)
	request := alidns.CreateDeleteSubDomainRecordsRequest()
	request.Scheme = "https"

	request.DomainName = domain
	request.RR = rr
	request.Type = recordType

	response, err = DDNSClient.DeleteSubDomainRecords(request)
	if err != nil {
		Logger.Printf("Failed to delete domain %s for %s type: %s\n", rr + "." + domain, recordType, err.Error())
	} else {
		Logger.Printf("Successfully remove %s record(s) from domain %s for %s type.\n", response.TotalCount, response.RR + "." + domain, recordType)
	}
	Debug.Printf("response is %#v\n", response)
	return response, err
}

// UpdateDomainRecord update domain
func UpdateDomainRecord(id string, rr string, ttl int, recordType string, value string) (response *alidns.UpdateDomainRecordResponse, err error) {
	Debug.Printf("updateRecord(%s, %s, %d, %s, %s)\n", id, rr, ttl, recordType, value)
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.RecordId = id
	request.RR = rr
	request.TTL = requests.NewInteger(ttl)
	request.Type = recordType
	request.Value = value

	response, err = DDNSClient.UpdateDomainRecord(request)
	if err != nil {
		Logger.Printf("Failed to update domain %s for %s type: %s\n", rr + "." + domain, recordType, err.Error())
	} else {
		Logger.Printf("Successfully update record from domain %s for %s type: %s\n", rr + "." + domain, recordType, response.RecordId)
	}
	Debug.Printf("response is %#v\n", response)
	return response, err
}

// AddDomainRecord add new domain
func AddDomainRecord(domain string, rr string, ttl int, recordType string, value string) (response *alidns.AddDomainRecordResponse, err error) {
	Debug.Printf("addRecord(%s, %s, %d, %s, %s)\n", domain, rr, ttl, recordType, value)
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"

	request.DomainName = domain
	request.RR = rr
	request.Type = recordType
	request.Value = value
	request.Line = line
	request.TTL = requests.NewInteger(ttl)

	response, err = DDNSClient.AddDomainRecord(request)
	if err != nil {
		Logger.Printf("Failed to add domain %s for %s type: %s\n", rr + "." + domain, recordType, err.Error())
	} else {
		Logger.Printf("Successfully add domain %s type record %s for %s at default line: %s.\n", recordType, value,  rr + "." + domain, response.RecordId)
	}
	Debug.Printf("response is %#v\n", response)
	return response, err
}
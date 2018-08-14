package core

import (
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"
	"net"
	"sort"
)

// InitCloudflare connects and validate Cloudflare credentials
func InitCloudflare(configFile *viper.Viper) (*cloudflare.API, *cloudflare.DNSRecord) {
	recordInfo := cloudflare.DNSRecord{
		ID:       "",
		Name:     configFile.GetString("DnsInfos.Name"),
		Type:     configFile.GetString("DnsInfos.Type"),
		Content:  "",
		TTL:      configFile.GetInt("DnsInfos.Ttl"),
		Proxied:  configFile.GetBool("DnsInfos.Proxied"),
		ZoneID:   configFile.GetString("CloudFlareApiInfos.ZoneId"),
		ZoneName: configFile.GetString("CloudFlareApiInfos.ZoneName"),
	}

	// Connect to Cloudflare
	log.Debugf("connecting to Cloudflare")
	api, err := cloudflare.New(configFile.GetString("CloudFlareApiInfos.Key"), configFile.GetString("CloudFlareApiInfos.Email"))
	if err != nil {
		log.Fatalf("Was not able to validate credentials to CloudFlare API: %v", err)
	}

	return api, &recordInfo
}

// UpdateDNSRecord add or remove DNS entry from the given DNS record
// record: node name in kubernetes
// recordContent: node name (dns record value)
// status: true=add/false=remove from DNS RR
func UpdateDNSRecord(api *cloudflare.API, record string, recordContent string, status bool, configFile *viper.Viper) bool {
	recordInfo := cloudflare.DNSRecord{
		ID:       "",
		Name:     configFile.GetString("DnsInfos.Name"),
		Type:     configFile.GetString("DnsInfos.Type"),
		Content:  recordContent,
		TTL:      configFile.GetInt("DnsInfos.Ttl"),
		Proxied:  configFile.GetBool("DnsInfos.Proxied"),
		ZoneID:   configFile.GetString("CloudFlareApiInfos.ZoneId"),
		ZoneName: configFile.GetString("CloudFlareApiInfos.ZoneName"),
	}
	// Todo: remove record? not useful
	fqdn := record + "." + recordInfo.ZoneName

	// Ensure content field is ok
	if recordInfo.Content == "nil" {
		log.Errorf("DNS content was not defined, skipping for host: %s", fqdn)
		return false
	}

	// Skip if record is already in the desired state
	recordResult, recordExist := GetDNSRecords(api, &recordInfo)
	if recordExist == status {
		log.Infof("Change detected, but no need to update current DNS record. Skipping for %s", fqdn)
		return false
	}

	// Make DNS change
	dnsRecord := fmt.Sprintf("%s -> %s (%s)", recordInfo.Name, recordInfo.Content, fqdn)
	if recordExist == true {
		recordInfo.ID = recordResult[0].ID
		result := deleteCurrentRecord(api, &recordInfo)
		if result == false {
			log.Errorf("Wasn't able to delete record: %s", dnsRecord)
			return false
		}
		log.Infof("Record DNS deleted: %s", dnsRecord)
	} else {
		result := createCurrentRecord(api, &recordInfo)
		if result == false {
			log.Error("Wasn't able to create record: %s", dnsRecord)
			return false
		}
		log.Infof("Record DNS created: %s", dnsRecord)
	}
	return true
}

// GetDNSRecords returns the list of records assigned to the round robin and a boolean saying if the record sent exist in the round robin
func GetDNSRecords(api *cloudflare.API, record *cloudflare.DNSRecord) ([]cloudflare.DNSRecord, bool) {
	// checking if record exists
	recordResult, err := api.DNSRecords(record.ZoneID, *record)
	if err != nil {
		log.Errorf("Error while trying to get %s record info: %v", record.Name, err)
		return recordResult, false
	}

	recordsSize := len(recordResult)
	if recordsSize == 0 {
		return recordResult, false
	}
	return recordResult, true
}

// GetDNSRecordValueIP perform a dns lookup and return corresponding IP
func GetDNSRecordValueIP(fqdn string, obj *v1.Node, configFile *viper.Viper) string {
	if configFile.GetString("GlobalConfig.UpdateDnsType") == "dns" {
		ipAddress, err := net.LookupHost(fqdn)
		if err != nil {
			log.Errorf("Can't find a DNS record for %s. It is advised to add server DNS record or filter it with labels", fqdn)
			return "nil"
		}
		return ipAddress[0]
	}
	return obj.Status.Addresses[0].Address
}

// GetIPFromDNS perform a dns lookup and return corresponding IP
func GetIPFromDNS(fqdn string) string {
	ipAddress, err := net.LookupHost(fqdn)
	if err != nil {
		log.Errorf("Can't find a DNS record for %s", fqdn)
		return "nil"
	}
	return ipAddress[0]
}

// IsDNSUpdateRequired validate if the current status of the round robin entries from CloudFlare API match the desired number of entries
// Return a boolean saying if we need to update current round robin configuration and the list of records assigned to the round robin
func IsDNSUpdateRequired(api *cloudflare.API, record *cloudflare.DNSRecord, configFile *viper.Viper) (bool, []cloudflare.DNSRecord) {
	// Get the current list of declared records for the round robin
	dnsList, _ := GetDNSRecords(api, record)
	currentRrDNSSize := len(dnsList)
	wantedRrDNSEntries := configFile.GetInt("GlobalConfig.WantedRrDnsEntries")

	if currentRrDNSSize == wantedRrDNSEntries {
		log.Debugf("current RR list update not required: current(%d)/wanted(%d)", currentRrDNSSize, wantedRrDNSEntries)
		return false, dnsList
	}
	log.Debugf("current RR list update required: current(%d)/wanted(%d)", currentRrDNSSize, wantedRrDNSEntries)
	return true, dnsList
}

// GetCurrentDNSRecordsList return current DNS records as list from Cloudflare API
func GetCurrentDNSRecordsList(api *cloudflare.API, record *cloudflare.DNSRecord, configFile *viper.Viper) ([]cloudflare.DNSRecord) {
	// Get the current list of declared records for the round robin
	dnsList, _ := GetDNSRecords(api, record)
	log.Infof("Current round robin list update required: current(%d)/wanted(%d)", len(dnsList), configFile.GetInt("GlobalConfig.WantedRrDnsEntries"))
	return dnsList
}

// UpdateRandomDNSEntries add random entries in the RR DNS from schedulable nodes
// k8sNodes: list of kubernetes nodes with their state
// currentRegisteredDns: list of current registered DNS in round robin
// action: 0=add / 1=remove
// numberOfIteration: how many add or remove to perfom
func UpdateRandomDNSEntries(api *cloudflare.API, configFile *viper.Viper, k8sNodes map[string]bool, currentRegisteredDNS []string, action int, numberOfIteration int) bool {
	counter := 0
	if action < 0 || action > 1 {
		log.Errorf("Bad DNS update action, doesn't understand what to do")
		return false
	}

	sort.Strings(currentRegisteredDNS)
	for nodeName, unschedulable := range k8sNodes {
		if counter >= numberOfIteration {
			return true
		}

		// Ensure node is not already in DNS and is schedulable
		if sort.SearchStrings(currentRegisteredDNS, nodeName) == action && unschedulable == false {
			fqdn := nodeName + "." + configFile.GetString("CloudFlareAPIInfos.ZoneName")
			updateReturnStatus := UpdateDNSRecord(api, nodeName, GetIPFromDNS(fqdn), true, configFile)
			if updateReturnStatus == true {
				counter++
			}
		}
	}
	return false
}

func deleteCurrentRecord(api *cloudflare.API, record *cloudflare.DNSRecord) bool {
	err := api.DeleteDNSRecord(record.ZoneID, record.ID)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func createCurrentRecord(api *cloudflare.API, record *cloudflare.DNSRecord) bool {
	_, err := api.CreateDNSRecord(record.ZoneID, *record)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

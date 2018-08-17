package core

import (
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
	"strconv"
	"github.com/spf13/viper"
)

func UpdateDnsRecord(record string, recordContent string, status bool, configFile *viper.Viper) {
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
	fqdn := record + "." + recordInfo.ZoneName

	// Ensure content field is ok
	if recordInfo.Content == "nil" {
		log.Errorf("DNS content was not defined, skipping for host: %s", fqdn)
	}

	// Connect to Cloudflare
	log.Debugf("requesting %s dns record to %s", record, strconv.FormatBool(status))
	cloudFlareApi := cloudFlareConnect(configFile.GetString("CloudFlareApiInfos.Key"), configFile.GetString("CloudFlareApiInfos.Email"))

	// Skip if record is already in the desired state
	recordResult, recordExist := checkDnsRecordExist(cloudFlareApi, &recordInfo)
	if recordExist == status {
		log.Infof("Change detected, but no need to update current DNS record. Skipping for %s", fqdn)
		return
	}

	// Make DNS change
	dnsRecord := fmt.Sprintf("%s -> %s (%s)", recordInfo.Name, recordInfo.Content, fqdn)
	if recordExist == true {
		recordInfo.ID = recordResult[0].ID
		result := deleteCurrentRecord(cloudFlareApi, &recordInfo)
		if result == false {
			log.Errorf("Wasn't able to delete record: %s", dnsRecord)
		} else {
			log.Infof("Record DNS deleted: %s", dnsRecord)
		}
	} else {
		result := createCurrentRecord(cloudFlareApi, &recordInfo)
		if result == false {
			log.Error("Wasn't able to create record: %s", dnsRecord)
		} else {
			log.Infof("Record DNS created: %s", dnsRecord)
		}
	}
}

func cloudFlareConnect(login string, key string) *cloudflare.API {
	api, err := cloudflare.New(login, key)
	if err != nil {
		log.Fatalf("Was not able to validate credentials to CloudFlare API: %s", err.Error())
	}
	return api
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

func checkDnsRecordExist(api *cloudflare.API, record *cloudflare.DNSRecord) ([]cloudflare.DNSRecord, bool) {
	recordResult, err := api.DNSRecords(record.ZoneID, *record)
	if err != nil || len(recordResult) == 0 {
		log.Debugf("checking record %s in DNS zone(%s): does not exist", record.Name, strconv.Itoa(len(recordResult)))
		return recordResult, false
	}
	log.Debugf("checking record %s: exists", record.Name)
	return recordResult, true
}

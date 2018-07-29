package core

import (
	"strconv"
	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
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
		log.Error("DNS content was not defined, skipping for host: " + fqdn)
	}

	// Connect to Cloudflare
	log.Debug("Requesting " + record + " dns record to " + strconv.FormatBool(status))
	cloudFlareApi := cloudFlareConnect(configFile.GetString("CloudFlareApiInfos.Key"), configFile.GetString("CloudFlareApiInfos.Email"))

	// Skip if record is already in the desired state
	recordResult, recordExist := checkDnsRecordExist(*cloudFlareApi, recordInfo)
	if recordExist == status {
		log.Info("Change detected, but no need to update current DNS record. Skipping for " + fqdn)
		return
	}

	// Make DNS change
	dnsRecord :=  recordInfo.Name + " -> " + recordInfo.Content + " (" + fqdn + ")"
	if recordExist == true {
		recordInfo.ID = recordResult[0].ID
		result := deleteCurrentRecord(*cloudFlareApi, recordInfo)
		if result == false {
			log.Error("Wasn't able to delete record: " + dnsRecord)
		}
		log.Info("Record DNS deleted: " + dnsRecord)
	} else {
		result := createCurrentRecord(*cloudFlareApi, recordInfo)
		if result == false {
			log.Error("Wasn't able to create record: " + dnsRecord)
		}
		log.Info("Record DNS created: " + dnsRecord)
	}
}

func cloudFlareConnect(login string, key string) *cloudflare.API {
	api, err := cloudflare.New(login, key)
	if err != nil {
		log.Error("Was not able to validate credentials to CloudFlare API")
		log.Fatal(err)
	}

	return api
}

func deleteCurrentRecord(api cloudflare.API, record cloudflare.DNSRecord) bool {
	err := api.DeleteDNSRecord(record.ZoneID, record.ID)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func createCurrentRecord(api cloudflare.API, record cloudflare.DNSRecord) bool {
	_, err := api.CreateDNSRecord(record.ZoneID, record)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func checkDnsRecordExist(api cloudflare.API, record cloudflare.DNSRecord) ([]cloudflare.DNSRecord, bool) {
	recordResult, err := api.DNSRecords(record.ZoneID, record)
	if err != nil || len(recordResult) == 0 {
		log.Debug("Checking record " + record.Name + " in DNS zone(" + strconv.Itoa(len(recordResult)) + "): does not exist")
		return recordResult, false
	}
	log.Debug("Checking record " + record.Name + ": exists")
	return recordResult, true
}
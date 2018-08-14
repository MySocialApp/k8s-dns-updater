package core

import (
	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Main is the core function
func Main() {
	log.Info("Starting k8s-dns-updater")
	// Load config file and connect to kubernetes cluster
	k8sConnect, yamlConfig := Init()
	// Connect to Cloudflare
	cloudFlareAPI, dnsRecord := InitCloudflare(yamlConfig)

	// Reassign at start to ensure everything is as expected
	dnsList := GetCurrentDNSRecordsList(cloudFlareAPI, dnsRecord, yamlConfig)
	ReassignDNSRrEntries(cloudFlareAPI, k8sConnect, yamlConfig, dnsRecord, dnsList)

	// Continuously Watch node changes
	WatchNodes(k8sConnect, cloudFlareAPI, yamlConfig)
	// Todo: regularely check if wanted number of dns == cloudflare list ?
}

// ReassignDNSRrEntries will ensure the number of wanted entries in the round robin is respected
func ReassignDNSRrEntries(api *cloudflare.API, k8sConnect *kubernetes.Clientset, configFile *viper.Viper, record *cloudflare.DNSRecord, dnsList []cloudflare.DNSRecord) {
	var dnsSchedulableNodes []string
	var dnsUnschedulableNodes []string
	var currentRegistredDNS []string
	var nodeName string

	// Get current node status from kubernetes
	k8sNodes := GetK8sNodesStatus(k8sConnect)

	// Compare DNS list with active nodes and select schedulable and unschedulable nodes declared
	for _, recordItem := range dnsList {
		// Define dns or node name
		nodeName = recordItem.Content
		if configFile.GetString("GlobalConfig.UpdateDnsType") == "dns" {
			nodeName = recordItem.Content + "." + configFile.GetString("CloudFlareAPIInfos.ZoneName")
		}
		currentRegistredDNS = append(currentRegistredDNS, nodeName)

		if k8sNodes[nodeName] {
			dnsSchedulableNodes = append(dnsSchedulableNodes, recordItem.Content)
		} else {
			dnsUnschedulableNodes = append(dnsUnschedulableNodes, recordItem.Content)
		}
	}

	// If there is less available entries than wanted -> add DNS entries in RR
	if iteration := configFile.GetInt("GlobalConfig.WantedRrDNSEntries") - len(dnsUnschedulableNodes) ; len(currentRegistredDNS) < iteration {
		log.Debugf("adding %d entries in the DNS", iteration)
		UpdateRandomDNSEntries(api, configFile, k8sNodes, currentRegistredDNS, 0, iteration)
	// If there is more available entries than wanted - unschedulable nodes -> remove DNS entries in RR
	} else if iteration := len(dnsSchedulableNodes) - len(dnsUnschedulableNodes) - configFile.GetInt("GlobalConfig.WantedRrDNSEntries") ; iteration > 0 {
		log.Debugf("removing %d entries in the DNS", iteration)
		UpdateRandomDNSEntries(api, configFile, k8sNodes, currentRegistredDNS, 1, iteration)
	} else {
		log.Debug("no dns update is required")
		return
	}

	// Clean unschedulable entries
	for _, nodeName := range dnsUnschedulableNodes {
		UpdateDNSRecord(api, nodeName, nodeName, false, configFile)
	}
}

// GetK8sNodesStatus return a hashmap of kubernetes node status
func GetK8sNodesStatus(k8sConnect *kubernetes.Clientset) map[string]bool {
	nodes := make(map[string]bool)

	nodeList, err := k8sConnect.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		log.Fatalf("Can't get nodes status from Kubernetes: %s", err)
	}

	for _, item := range nodeList.Items {
		nodes[item.ObjectMeta.Name] = item.Spec.Unschedulable
	}

	return nodes
}
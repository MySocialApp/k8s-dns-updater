package core

import (
	"k8s.io/client-go/kubernetes"
	"strconv"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"time"
	"github.com/spf13/viper"
	"net"
)

func WatchNodes(clientSet *kubernetes.Clientset, configFile *viper.Viper) {
	var nodeStatus string
	var nodeStatusBool bool
	watchlist := cache.NewListWatchFromClient(clientSet.CoreV1().RESTClient(), "nodes", v1.NamespaceAll, fields.Everything())

	informer := cache.NewSharedIndexInformer(
		watchlist,
		&v1.Node{},
		time.Second,
		cache.Indexers{},)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Show node status at start
		AddFunc: func(obj interface{}) {
			log.Info("First check: " +  obj.(*v1.Node).ObjectMeta.Name + " unschedulable status node is " + strconv.FormatBool(obj.(*v1.Node).Spec.Unschedulable))
		},

		// Show deleted node
		DeleteFunc: func(obj interface{}) {
			log.Info("Deleted node: " + obj.(*v1.Node).ObjectMeta.Name)
		},

		// Detect changes
		UpdateFunc: func(oldObj, currentObj interface{}) {
			currentNodeStatus := strconv.FormatBool(currentObj.(*v1.Node).Spec.Unschedulable)
			oldNodeStatus := strconv.FormatBool(oldObj.(*v1.Node).Spec.Unschedulable)

			if currentNodeStatus != oldNodeStatus {
				record := currentObj.(*v1.Node).ObjectMeta.Name
				fqdn := record + "." + configFile.GetString("CloudFlareApiInfos.ZoneName")
				recordContent := getRecordValueIp(fqdn, currentObj, configFile)

				nodeStatus = "enabled"
				nodeStatusBool = true
				if currentObj.(*v1.Node).Spec.Unschedulable {
					nodeStatus = "disabled"
					nodeStatusBool = false
				}

				log.Info("Scheduling node " + record + " changed to " + nodeStatus)
				UpdateDnsRecord(record, recordContent, nodeStatusBool, configFile)
			}
		},
	})

	stop := make(chan struct{})
	go informer.Run(stop)

	for{
		time.Sleep(time.Second)
	}
}

func getRecordValueIp(fqdn string, obj interface{}, configFile *viper.Viper) string {
	if configFile.GetString("GlobalConfig.UpdateDnsType") == "dns" {
		ipAddress, err := net.LookupHost(fqdn)
		if err != nil {
			log.Error("Can't find a DNS record for " + fqdn)
			return "nil"
		}
		return ipAddress[0]
	} else {
		return obj.(*v1.Node).Status.Addresses[0].Address
	}
	return "nil"
}
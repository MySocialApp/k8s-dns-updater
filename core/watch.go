package core

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"net"
	"strconv"
	"time"
)

func WatchNodes(clientSet *kubernetes.Clientset, configFile *viper.Viper) {
	var nodeStatus string
	var nodeStatusBool bool
	watchlist := cache.NewListWatchFromClient(clientSet.CoreV1().RESTClient(), "nodes", v1.NamespaceAll, fields.Everything())

	informer := cache.NewSharedIndexInformer(
		watchlist,
		&v1.Node{},
		time.Second,
		cache.Indexers{})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Show node status at start
		AddFunc: func(obj interface{}) {
			if node, ok := obj.(*v1.Node); ok {
				log.Infof("Initial check: %s unschedulable status node is %s", node.ObjectMeta.Name, strconv.FormatBool(node.Spec.Unschedulable))
			}
		},

		// Show deleted node
		DeleteFunc: func(obj interface{}) {
			log.Infof("Deleted node: %s", obj.(*v1.Node).ObjectMeta.Name)
		},

		// Detect changes
		UpdateFunc: func(oldObj interface{}, currentObj interface{}) {
			if node, ok := currentObj.(*v1.Node); ok {
				currentNodeStatus := strconv.FormatBool(node.Spec.Unschedulable)
				oldNodeStatus := strconv.FormatBool(oldObj.(*v1.Node).Spec.Unschedulable)

				if currentNodeStatus != oldNodeStatus {
					record := node.ObjectMeta.Name
					fqdn := record + "." + configFile.GetString("CloudFlareApiInfos.ZoneName")
					recordContent := getDnsRecordValueIp(fqdn, node, configFile)

					nodeStatus = "enabled"
					nodeStatusBool = true
					if node.Spec.Unschedulable {
						nodeStatus = "disabled"
						nodeStatusBool = false
					}

					log.Infof("Scheduling node %s changed to %s", record, nodeStatus)
					UpdateDnsRecord(record, recordContent, nodeStatusBool, configFile)
				}
			}
		},
	})

	stop := make(chan struct{})
	go informer.Run(stop)

	<-stop
}

func getDnsRecordValueIp(fqdn string, obj *v1.Node, configFile *viper.Viper) string {
	if configFile.GetString("GlobalConfig.UpdateDnsType") == "dns" {
		ipAddress, err := net.LookupHost(fqdn)
		if err != nil {
			log.Errorf("Can't find a DNS record for %s", fqdn)
			return "nil"
		}
		return ipAddress[0]
	} else {
		return obj.Status.Addresses[0].Address
	}
	return "nil"
}

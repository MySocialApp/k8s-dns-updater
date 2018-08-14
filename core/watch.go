package core

import (
	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"strconv"
	"time"
)

// WatchNodes is watching kubernetes nodes changes and update DNS accordingly
func WatchNodes(clientSet *kubernetes.Clientset, cloudFlareAPI *cloudflare.API, configFile *viper.Viper) {
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
				log.Infof("Initial k8s node check: %s unschedulable status node is %s", node.ObjectMeta.Name, strconv.FormatBool(node.Spec.Unschedulable))
			}
		},

		// Show deleted node
		DeleteFunc: func(obj interface{}) {
			if node, ok := obj.(*v1.Node); ok {
				log.Infof("Deleted node: %s", node.ObjectMeta.Name)
			}
		},

		// Detect changes
		UpdateFunc: func(oldObj interface{}, currentObj interface{}) {
			if node, ok := currentObj.(*v1.Node); ok {
				currentNodeStatus := strconv.FormatBool(node.Spec.Unschedulable)
				oldNodeStatus := strconv.FormatBool(oldObj.(*v1.Node).Spec.Unschedulable)

				if currentNodeStatus != oldNodeStatus {
					record := node.ObjectMeta.Name
					fqdn := record + "." + configFile.GetString("CloudFlareApiInfos.ZoneName")
					recordContent := GetDNSRecordValueIP(fqdn, node, configFile)

					nodeStatus = "enabled"
					nodeStatusBool = true
					if node.Spec.Unschedulable {
						nodeStatus = "disabled"
						nodeStatusBool = false
					}

					log.Infof("Scheduling node %s changed to %s", record, nodeStatus)
					UpdateDNSRecord(cloudFlareAPI, record, recordContent, nodeStatusBool, configFile)
					// Todo: ReassignDnsRrEntries(cloudFlareApi) ?
				}
			}
		},
	})

	stop := make(chan struct{})
	go informer.Run(stop)

	<-stop
}
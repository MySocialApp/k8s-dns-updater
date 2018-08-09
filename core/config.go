package core

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
			)

type Configuration struct {
	GlobalConfig       GlobalConfig
	DnsInfos           DnsInfos
	CloudFlareApiInfos CloudFlareApiInfos
}

type GlobalConfig struct {
	UpdateDnsType string
	MaxDnsEntries int
}

type DnsInfos struct {
	Name    string
	Type    string
	Ttl     int
	Proxied bool
}

type CloudFlareApiInfos struct {
	ZoneId   string
	ZoneName string
	Email    string
	Key      string
}

var configuration Configuration

func Init() (*kubernetes.Clientset, *viper.Viper) {
	return connectToKubernetes(getK8sConfig()), getConfigFromYamlFile()
}

func getK8sConfig() *rest.Config {
	var config *rest.Config
	var err error

	// Try Kubeconfig first
	config, err = getK8sConfigFromKubeconfig()
	if err == nil {
		return config
	}

	// Else in cluster config
	log.Debug("Can't get kubeconfig information, trying to connect if is inside the cluster")
	config, err = getK8sConfigInCluster()
	if err != nil {
		log.Fatalf("Can't connect to Kubernetes: %s", err.Error())
	}

	return config
}

func getK8sConfigFromKubeconfig() (*rest.Config, error) {
	var kubeconfig string

	// Load kubeconfig file
	log.Debug("Getting Kubernetes config from kubeconfig")
	if home := homeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func getK8sConfigInCluster() (*rest.Config, error) {
	return rest.InClusterConfig()
}

func getConfigFromYamlFile() *viper.Viper {
	config := viper.New()

	currentPath, _ := filepath.Abs("./")
	config.AddConfigPath(currentPath)
	config.AddConfigPath("/etc/k8s-dns-updater")
	config.SetConfigName("config")

	if err := config.ReadInConfig(); err != nil {
		log.Debug(err)
		log.Fatalf("Config file config.yaml not found in %s or /etc/k8s-dns-updater", currentPath)
	}
	log.Debug("Using config file: ", config.ConfigFileUsed())

	return config
}

func connectToKubernetes(config *rest.Config) *kubernetes.Clientset {
	// Create the clientset connection to Kubernetes
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error while creating the connection to Kubernetes: %s", err.Error())
	}

	return clientset
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

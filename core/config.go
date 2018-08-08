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
	return connectToKubernetes(getK8sConfigFromKubeconfig()), getConfigFromYamlFile()
}

func getK8sConfigFromKubeconfig() *rest.Config {
	var kubeconfig string

	// Load kubeconfig file
	log.Debug("Getting Kubernetes config from kubeconfig")
	if home := homeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error while looking at Kubernetes context: %s", err.Error())
	}

	return config
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

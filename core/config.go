package core

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

// Configuration represent the complete configuration
type Configuration struct {
	GlobalConfig       GlobalConfig
	DNSInfos           DNSInfos
	CloudFlareAPIInfos CloudFlareAPIInfos
}

// GlobalConfig is application related
type GlobalConfig struct {
	UpdateDNSType      string
	WantedRrDNSEntries int
}

// DNSInfos for what you expect on your DNS record
type DNSInfos struct {
	Name    string
	Type    string
	TTL     int
	Proxied bool
}

// CloudFlareAPIInfos are CloudFlare connection API info
type CloudFlareAPIInfos struct {
	ZoneID   string
	ZoneName string
	Email    string
	Key      string
}

var configuration Configuration

// Init returns Kubernetes connection and yaml config
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
	log.Debug("can't get kubeconfig information, trying to connect if is inside the cluster")
	config, err = getK8sConfigInCluster()
	if err != nil {
		log.Fatalf("Can't connect to Kubernetes: %s", err.Error())
	}

	return config
}

func getK8sConfigFromKubeconfig() (*rest.Config, error) {
	// Load kubeconfig file
	log.Debug("getting Kubernetes config from kubeconfig")
	kubeconfig := ""
	if home := homeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

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
		log.Fatalf("config file config.yaml not found in %s or /etc/k8s-dns-updater", currentPath)
	}
	log.Debug("using config file: ", config.ConfigFileUsed())

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

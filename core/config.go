package core

import (
	"flag"
	"os"
	"path/filepath"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	)

type Configuration struct {
	GlobalConfig GlobalConfig
	DnsInfos DnsInfos
	CloudFlareApiInfos CloudFlareApiInfos
}

type GlobalConfig struct {
	UpdateDnsType string
	MaxDnsEntries int
}

type DnsInfos struct {
	Name string
	Type string
	Ttl int
	Proxied bool
}

type CloudFlareApiInfos struct {
	ZoneId string
	ZoneName string
	Email string
	Key string
}

var configuration Configuration

func Init() (*kubernetes.Clientset, *viper.Viper) {
	// Load Yaml configuration from file
	configFile := *getConfigFromYamlFile()

	// Get Kubeconfig info
	kubeConfig := *getK8sConfigFromKubeconfig()

	// Connect to Kubernetes
	clientSet := *connectToKubernetes(&kubeConfig)

	return &clientSet, &configFile
}

func getK8sConfigFromKubeconfig() *rest.Config {
	var kubeconfig *string

	// Load kubeconfig file
	log.Debug("Getting Kubernetes config from kubeconfig")
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Error("Error while looking at Kubernetes context")
		panic(err.Error())
	}

	return config
}

func getConfigFromYamlFile() *viper.Viper {
	config := viper.New()

	currentPath, _ := filepath.Abs("./")
	config.AddConfigPath(currentPath)
	config.SetConfigName("config")

	if err := config.ReadInConfig() ; err != nil {
		log.Debug(err)
		log.Fatal("Config file not found: " + currentPath + "/config.yaml")
	}
	log.Debug("Using config file: ", config.ConfigFileUsed())

	return config
}

func connectToKubernetes(config *rest.Config) *kubernetes.Clientset {
	// Create the clientset connection to Kubernetes
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error("Error while creating the connection to Kubernetes")
		panic(err.Error())
	}

	return clientset
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

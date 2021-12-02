package main

import (
	"flag"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	log "github.com/sirupsen/logrus"

	"github.com/joeyloman/kube-fip-operator/pkg/app"
	"github.com/joeyloman/kube-fip-operator/pkg/file"

	kubefipclientset "github.com/joeyloman/kube-fip-operator/pkg/generated/clientset/versioned"
)

var progname string = "kube-fip-operator"

func init() {
	// Log as JSON instead of the default ASCII formatter.
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	homedir := os.Getenv("HOME")
	kubeconfig_file := filepath.Join(homedir, ".kube", "config")

	log.Infof("(main) starting %s ..", progname)

	var config *rest.Config = nil
	if file.FileExists(kubeconfig_file) {
		// uses kubeconfig
		kubeconfig := flag.String("kubeconfig", kubeconfig_file, "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		config_kube, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		config = config_kube
	} else {
		// creates the in-cluster config
		config_rest, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = config_rest
	}

	// create the default k8s clientset
	k8s_clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// create the kubefip clientset
	kubefip_clientset, err := kubefipclientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	app.Run(kubefip_clientset, k8s_clientset)
}

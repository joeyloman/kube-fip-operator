package app

import (
	"github.com/joeyloman/kube-fip-operator/pkg/config"
	kubefipclientset "github.com/joeyloman/kube-fip-operator/pkg/generated/clientset/versioned"
	"github.com/joeyloman/kube-fip-operator/pkg/kubefip"
	"github.com/joeyloman/kube-fip-operator/pkg/metrics"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

func Run(kubefip_clientset *kubefipclientset.Clientset, k8s_clientset *kubernetes.Clientset) {
	kubefipConfigmap, err := config.GetKubefipConfigmap(k8s_clientset)
	if err != nil {
		log.Errorf("(Run) %s", err)
		log.Debugf("(Run) applying defaults..")
	}

	// parse all config options
	kubefipConfig := config.ParseKubfipConfigMap(kubefipConfigmap)

	// update the loglevel
	updateLoglevel(&kubefipConfig)

	// init all metrics as a goroutine (separate thread)
	go metrics.InitMetrics(kubefipConfig.MetricsPort)

	// create an array with all the FipRange objects
	if err := kubefip.GatherAllFipRanges(kubefip_clientset); err != nil {
		log.Errorf("(Run) error gathering all FipRanges: %s", err.Error())
	}

	for i := 0; i < len(kubefip.AllFipRanges); i++ {
		log.Infof("(Run) stored fiprange name [%s] and cidr [%s]", kubefip.AllFipRanges[i].ObjectMeta.Name, kubefip.AllFipRanges[i].Spec.IPRange)
	}

	// create an array with all the Fip objects
	if err := kubefip.GatherAllFips(k8s_clientset, kubefip_clientset); err != nil {
		log.Errorf("(Run) error gathering all fips: %s", err.Error())
	}

	for i := 0; i < len(kubefip.AllFips); i++ {
		log.Infof("(Run) stored fip name [%s/%s] and ipaddress [%s]", kubefip.AllFips[i].ObjectMeta.Namespace, kubefip.AllFips[i].ObjectMeta.Name,
			kubefip.AllFips[i].Spec.IPAddress)
	}

	// init the ipam modules
	kubefip.InitIpam()

	// store all ip ranges as a prefix object in the ipam object
	kubefip.CreateIpamPrefixesFromFipRanges()

	// put all the existing fips objects in the ipam object
	kubefip.StoreAllocatedIpsInIpamPrefixes(kubefip_clientset)

	// start the maintaining of the kubevip configs
	startManageKubevip(k8s_clientset, &kubefipConfig)

	// start watching the namespace and secret events
	watchEvents(kubefip_clientset, k8s_clientset, &kubefipConfig)
}

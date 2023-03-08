package app

import (
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	log "github.com/sirupsen/logrus"

	KubefipV1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	"github.com/joeyloman/kube-fip-operator/pkg/config"
	kubefipclientset "github.com/joeyloman/kube-fip-operator/pkg/generated/clientset/versioned"
	"github.com/joeyloman/kube-fip-operator/pkg/kubefip"
	"github.com/joeyloman/kube-fip-operator/pkg/metrics"

	corev1 "k8s.io/api/core/v1"
)

func watchEvents(kubefip_clientset *kubefipclientset.Clientset, k8s_clientset *kubernetes.Clientset, kubefipConfig *config.KubefipConfigStruct) {
	var watchEventTimeout int = 30 // in seconds (skip all events for the first 30 secs)

	log.Infof("(watchEvents) start watching the floatingipaddress, floatingiprange, namespace and configmap events ..")

	// toggle watchEventsActivated after 10 secs
	var watchEventsActivated bool = false
	time.AfterFunc(time.Duration(watchEventTimeout)*time.Second, func() { watchEventsActivated = true })

	// do the eventwatch stuff for fips
	watchlistFips := cache.NewListWatchFromClient(kubefip_clientset.KubefipV1().RESTClient(), "floatingips", corev1.NamespaceAll,
		fields.Everything())

	_, controllerFips := cache.NewInformer(
		watchlistFips,
		&KubefipV1.FloatingIP{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Debugf("(watchFipEvents) entering the eventwatch AddFunc ..")

				if watchEventsActivated {
					// allocate the new Fip
					if err := kubefip.AllocateFip(obj.(*KubefipV1.FloatingIP), kubefip_clientset); err != nil {
						log.Errorf("(watchFipEvents) error allocating fip: %s", err.Error())
					}
				} else {
					log.Debugf("(watchFipEvents) not activated yet, object action not executed")
				}
			},
			DeleteFunc: func(obj interface{}) {
				log.Debugf("(watchFipEvents) entering the eventwatch DeleteFunc ..")

				if watchEventsActivated {
					// remove the Fip
					if err := kubefip.RemoveFip(obj.(*KubefipV1.FloatingIP)); err != nil {
						log.Errorf("(watchFipEvents) error removing fip: %s", err.Error())
					}

					// get the harvester clustername from the FipRange object (because the cluster and related objects are already gone from here)
					harvesterClusterName, err := getHarvesterClusterNameFromFipRange(obj.(*KubefipV1.FloatingIP), kubefip_clientset)
					if err != nil {
						log.Errorf("(watchFipEvents) error cannot get harvester clustername for fip: [%s]: %s",
							obj.(*KubefipV1.FloatingIP).ObjectMeta.Name, err.Error())
					}

					// add the cluster name and harvester cluster name to the metrics cleanup queue
					metrics.AddClusterToMetricsCleanupQueue(obj.(*KubefipV1.FloatingIP).ObjectMeta.Annotations["clustername"], harvesterClusterName)
				} else {
					log.Debugf("(watchFipEvents) not activated yet, object action not executed")
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				log.Debugf("(watchFipEvents) entering the eventwatch UpdateFunc ..")

				if watchEventsActivated {
					// update the Fip
					if err := kubefip.UpdateFip(oldObj.(*KubefipV1.FloatingIP), newObj.(*KubefipV1.FloatingIP), kubefip_clientset); err != nil {
						log.Errorf("(watchFipEvents) error removing fip: %s", err.Error())
					}
				} else {
					log.Debugf("(watchFipEvents) not activated yet, object action not executed")
				}
			},
		},
	)

	// do the eventwatch stuff for fipranges
	watchlistFipRanges := cache.NewListWatchFromClient(kubefip_clientset.KubefipV1().RESTClient(), "floatingipranges", corev1.NamespaceAll,
		fields.Everything())

	_, controllerFipRanges := cache.NewInformer(
		watchlistFipRanges,
		&KubefipV1.FloatingIPRange{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Debugf("(watchFipRangeEvents) entering the eventwatch AddFunc ..")

				if watchEventsActivated {
					// allocate the new FipRange
					if err := kubefip.AllocateFipRange(obj.(*KubefipV1.FloatingIPRange)); err != nil {
						log.Errorf("(watchFipRangeEvents) error allocating fiprange: %s", err.Error())
					}
				} else {
					log.Debugf("(watchFipEvents) not activated yet, object action not executed")
				}
			},
			DeleteFunc: func(obj interface{}) {
				log.Debugf("(watchFipRangeEvents) entering the eventwatch DeleteFunc ..")

				if watchEventsActivated {
					// remove the Fip
					if err := kubefip.RemoveFipRange(obj.(*KubefipV1.FloatingIPRange)); err != nil {
						log.Errorf("(watchFipRangeEvents) error removing fiprange: %s", err.Error())
					}
				} else {
					log.Debugf("(watchFipEvents) not activated yet, object action not executed")
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				log.Debugf("(watchFipRangeEvents) entering the eventwatch UpdateFunc ..")

				if watchEventsActivated {
					// update the Fip
					if err := kubefip.UpdateFipRange(oldObj.(*KubefipV1.FloatingIPRange), newObj.(*KubefipV1.FloatingIPRange)); err != nil {
						log.Errorf("(watchFipRangeEvents) error removing fiprange: %s", err.Error())
					}
				} else {
					log.Debugf("(watchFipEvents) not activated yet, object action not executed")
				}
			},
		},
	)

	// do the eventwatch stuff for namespaces so we can detect new clusters
	watchlistNamespaces := cache.NewListWatchFromClient(k8s_clientset.CoreV1().RESTClient(), "namespaces", corev1.NamespaceAll,
		fields.Everything())

	_, controllerNamespaces := cache.NewInformer(
		watchlistNamespaces,
		&corev1.Namespace{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Debugf("(watchNamespaceEvents) Entering the eventwatch AddFunc ..")

				if watchEventsActivated {
					// check if the new namespace is a cluster
					checkNewNamespace(obj.(*corev1.Namespace), kubefip_clientset, k8s_clientset)
				} else {
					log.Debugf("(watchNamespaceEvents) not activated yet, object action not executed")
				}
			},
		},
	)

	// do the eventwatch stuff for configmaps so we can detect new clusters
	watchlistConfigmaps := cache.NewListWatchFromClient(k8s_clientset.CoreV1().RESTClient(), "configmaps", "kube-fip", fields.Everything())

	_, controllerConfigmaps := cache.NewInformer(
		watchlistConfigmaps,
		&corev1.ConfigMap{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Debugf("(watchConfigmapEvents) entering the eventwatch AddFunc ..")

				if watchEventsActivated {
					if obj.(*corev1.ConfigMap).ObjectMeta.Name == "kube-fip-config" {
						log.Debugf("(watchConfigmapEvents) new kube-fip-config configmap found")

						// register the old operateGuestClusterInterval value
						oldOperateGuestClusterInterval := kubefipConfig.OperateGuestClusterInterval

						// parse the new configmap
						*kubefipConfig = config.ParseKubfipConfigMap(obj.(*corev1.ConfigMap))

						// update the loglevel
						updateLoglevel(kubefipConfig)

						// restart the operateTicker when the interval has changed
						restartManageKubevip(k8s_clientset, kubefipConfig, oldOperateGuestClusterInterval)
					}
				} else {
					log.Debugf("(watchConfigmapEvents) not activated yet, object action not executed")
				}
			},
			DeleteFunc: func(obj interface{}) {
				log.Debugf("(watchConfigmapEvents) entering the eventwatch DeleteFunc ..")

				if watchEventsActivated {
					if obj.(*corev1.ConfigMap).ObjectMeta.Name == "kube-fip-config" {
						log.Debugf("(watchConfigmapEvents) kube-fip-config configmap deleted")

						// register the old operateGuestClusterInterval value
						oldOperateGuestClusterInterval := kubefipConfig.OperateGuestClusterInterval

						// parse the new configmap
						*kubefipConfig = config.ParseKubfipConfigMap(nil)

						// update the loglevel
						updateLoglevel(kubefipConfig)

						// restart the operateTicker when the interval has changed
						restartManageKubevip(k8s_clientset, kubefipConfig, oldOperateGuestClusterInterval)
					}
				} else {
					log.Debugf("(watchConfigmapEvents) not activated yet, object action not executed")
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				log.Debugf("(watchConfigmapEvents) entering the eventwatch UpdateFunc ..")

				if watchEventsActivated {
					if newObj.(*corev1.ConfigMap).ObjectMeta.Name == "kube-fip-config" {
						log.Debugf("(watchConfigmapEvents) kube-fip-config configmap updated")

						// register the old operateGuestClusterInterval value
						oldOperateGuestClusterInterval := kubefipConfig.OperateGuestClusterInterval

						// parse the new configmap
						*kubefipConfig = config.ParseKubfipConfigMap(newObj.(*corev1.ConfigMap))

						// update the loglevel
						updateLoglevel(kubefipConfig)

						// restart the operateTicker when the interval has changed
						restartManageKubevip(k8s_clientset, kubefipConfig, oldOperateGuestClusterInterval)
					}
				} else {
					log.Debugf("(watchConfigmapEvents) not activated yet, object action not executed")
				}
			},
		},
	)

	stop := make(chan struct{})
	defer close(stop)
	go controllerFips.Run(stop)
	go controllerFipRanges.Run(stop)
	go controllerNamespaces.Run(stop)
	go controllerConfigmaps.Run(stop)

	for {
		time.Sleep(time.Second)
	}
}

package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	KubefipV1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	"github.com/joeyloman/kube-fip-operator/pkg/config"
	"github.com/joeyloman/kube-fip-operator/pkg/configmap"
	"github.com/joeyloman/kube-fip-operator/pkg/kubefip"
	"github.com/joeyloman/kube-fip-operator/pkg/metrics"

	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	operateTicker        *time.Ticker
	metricsCleanupTicker *time.Ticker

	updateMetrics     bool = true
	dontUpdateMetrics bool = false
)

func installKubevipCloudproviderInGuestCluster(kubeconfig []byte, kubefipConfig *config.KubefipConfigStruct, fip KubefipV1.FloatingIP) (bool, error) {
	var err error

	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return updateMetrics, err
	}

	opt := &helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        kubefipConfig.KubevipNamespace,
			RepositoryCache:  "/tmp/.helmcache-kube-vip",
			RepositoryConfig: "/tmp/.helmrepo-kube-vip",
			Debug:            true,
			Linting:          true,
		},
		RestConfig: config,
	}

	helmClient, err := helmclient.NewClientFromRestConf(opt)
	if err != nil {
		return updateMetrics, err
	}

	kubevipCoudproviderReleaseCheck, err := helmClient.GetRelease("kube-vip-cloud-provider")
	if err != nil {
		if err.Error() == "release: not found" {
			// now we can install it
			log.Infof("(installKubevipCloudproviderInGuestCluster) kube-vip-cloud-provider release not found in guest cluster [%s], trying to install it",
				fip.ObjectMeta.Annotations["clustername"])
		} else {
			return updateMetrics, err
		}
	}

	if kubevipCoudproviderReleaseCheck != nil {
		log.Debugf("(installKubevipCloudproviderInGuestCluster) helm chart already found in guest cluster [%s]: %s",
			fip.ObjectMeta.Annotations["clustername"], kubevipCoudproviderReleaseCheck.Name)

		// if not in update mode, don't update kube-vip
		if !kubefipConfig.KubevipUpdate {
			return dontUpdateMetrics, err
		}
	}

	chartRepo := repo.Entry{
		Name: "kube-vip",
		URL:  kubefipConfig.KubevipChartRepoUrl,
	}

	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		return updateMetrics, err
	}

	chartSpecKubevipCloudprovider := helmclient.ChartSpec{
		ReleaseName: "kube-vip-cloud-provider",
		ChartName:   "kube-vip/kube-vip-cloud-provider",
		Namespace:   kubefipConfig.KubevipNamespace,
		ValuesYaml:  kubefipConfig.KubevipCloudProviderChartValues,
		Wait:        false,
	}

	kubevipCloudproviderRelease, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpecKubevipCloudprovider, nil)
	if err != nil {
		return updateMetrics, err
	}

	log.Debugf("(installKubevipCloudproviderInGuestCluster) returned kube-vip-cloud-provider helm release manifest: %s",
		kubevipCloudproviderRelease.Manifest)

	log.Infof("(installKubevipCloudproviderInGuestCluster) kube-vip-cloud-provider helm chart installed successfully in guest cluster [%s]",
		fip.ObjectMeta.Annotations["clustername"])

	return updateMetrics, err
}

func installKubevipInGuestCluster(kubeconfig []byte, kubefipConfig *config.KubefipConfigStruct, fip KubefipV1.FloatingIP) (bool, error) {
	var err error

	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return updateMetrics, err
	}

	opt := &helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        kubefipConfig.KubevipNamespace,
			RepositoryCache:  "/tmp/.helmcache-kube-vip",
			RepositoryConfig: "/tmp/.helmrepo-kube-vip",
			Debug:            true,
			Linting:          true,
		},
		RestConfig: config,
	}

	helmClient, err := helmclient.NewClientFromRestConf(opt)
	if err != nil {
		return updateMetrics, err
	}

	kubevipReleaseCheck, err := helmClient.GetRelease("kube-vip")
	if err != nil {
		if err.Error() == "release: not found" {
			// now we can install it
			log.Infof("(installKubevipInGuestCluster) kube-vip release not found in guest cluster [%s], trying to install it",
				fip.ObjectMeta.Annotations["clustername"])
		} else {
			return updateMetrics, err
		}
	}

	if kubevipReleaseCheck != nil {
		log.Debugf("(installKubevipInGuestCluster) helm chart already found in guest cluster [%s]: %s",
			fip.ObjectMeta.Annotations["clustername"], kubevipReleaseCheck.Name)

		// if not in update mode, don't update kube-vip
		if !kubefipConfig.KubevipUpdate {
			return dontUpdateMetrics, err
		}
	}

	chartRepo := repo.Entry{
		Name: "kube-vip",
		URL:  kubefipConfig.KubevipChartRepoUrl,
	}

	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		return updateMetrics, err
	}

	chartSpecKubevip := helmclient.ChartSpec{
		ReleaseName: "kube-vip",
		ChartName:   "kube-vip/kube-vip",
		Namespace:   kubefipConfig.KubevipNamespace,
		ValuesYaml:  kubefipConfig.KubevipChartValues,
		Wait:        false,
	}

	kubevipRelease, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpecKubevip, nil)
	if err != nil {
		return updateMetrics, err
	}

	log.Debugf("(installKubevipInGuestCluster) returned kube-vip helm release manifest: %s",
		kubevipRelease.Manifest)

	log.Infof("(installKubevipInGuestCluster) kube-vip helm chart installed successfully in guest cluster [%s]",
		fip.ObjectMeta.Annotations["clustername"])

	return updateMetrics, err
}

func createOrUpdateKubevipConfigmapInGuestCluster(kubeconfig []byte, kubefipConfig *config.KubefipConfigStruct, fip KubefipV1.FloatingIP) (bool, error) {
	var kubevipConfigMapName string = "kubevip"
	var configMapExists bool = false
	var err error

	log.Debugf("(createKubevipConfigmapInGuestCluster) start connection to guest cluster")

	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return updateMetrics, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return updateMetrics, err
	}

	// list the configmaps in kube-system
	cmList, err := clientset.CoreV1().ConfigMaps(kubefipConfig.KubevipNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return updateMetrics, err
	}

	// check if the kubevip configmap already exists
	for _, cm := range cmList.Items {
		if cm.Name == kubevipConfigMapName {
			log.Debugf("(createKubevipConfigmapInGuestCluster) configmap [%s/%s] already exists in guest cluster [%s]",
				kubefipConfig.KubevipNamespace, kubevipConfigMapName, fip.ObjectMeta.Annotations["clustername"])

			configMapExists = true
			break
		}
	}

	if !configMapExists {
		// generating the new configmap
		newConfigMap := configmap.NewKubevipConfigmap(&fip, kubevipConfigMapName, kubefipConfig.KubevipNamespace)

		// creating the new configmap
		cmCreateObj, err := clientset.CoreV1().ConfigMaps(kubefipConfig.KubevipNamespace).Create(context.TODO(), &newConfigMap, metav1.CreateOptions{})
		if err != nil {
			errMsg := fmt.Sprintf("error creating kubevip configmap [%s/%s] in guest cluster [%s]: %s",
				kubefipConfig.KubevipNamespace, kubevipConfigMapName, fip.ObjectMeta.Annotations["clustername"], err.Error())
			return updateMetrics, errors.New(errMsg)
		}
		log.Tracef("(createKubevipConfigmapInGuestCluster) configmap obj created: [%s]", cmCreateObj)

		log.Infof("(createKubevipConfigmapInGuestCluster) successfully created configmap [%s/%s] in guest cluster [%s]",
			kubefipConfig.KubevipNamespace, kubevipConfigMapName, fip.ObjectMeta.Annotations["clustername"])
	} else {
		forceUpdate, err := strconv.ParseBool(fip.ObjectMeta.Annotations["updateConfigMap"])
		if err != nil {
			log.Debugf("(createKubevipConfigmapInGuestCluster) forceUpdate annotation error: %s", err)
		}

		if forceUpdate {
			// generating the new configmap
			newConfigMap := configmap.NewKubevipConfigmap(&fip, kubevipConfigMapName, kubefipConfig.KubevipNamespace)

			// updating the existing configmap
			cmUpdateObj, err := clientset.CoreV1().ConfigMaps(kubefipConfig.KubevipNamespace).Update(context.TODO(), &newConfigMap, metav1.UpdateOptions{})
			if err != nil {
				errMsg := fmt.Sprintf("error updating kubevip configmap [%s/%s] in guest cluster [%s]: %s", kubefipConfig.KubevipNamespace,
					kubevipConfigMapName, fip.ObjectMeta.Annotations["clustername"], err.Error())
				return updateMetrics, errors.New(errMsg)
			}
			log.Tracef("(createKubevipConfigmapInGuestCluster) configmap obj updated: [%s]", cmUpdateObj)

			log.Debugf("(createKubevipConfigmapInGuestCluster) successfully updated configmap [%s/%s] in guest cluster [%s]",
				kubefipConfig.KubevipNamespace, kubevipConfigMapName, fip.ObjectMeta.Annotations["clustername"])
		} else {
			return dontUpdateMetrics, err
		}
	}

	return updateMetrics, err
}

func testGuestClusterConnection(kubeconfig []byte) error {
	var err error

	log.Debugf("(testGuestClusterConnection) start checking the connection to the guest cluster")

	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	return nil
}

func getGuestClusterKubeconfig(clientset *kubernetes.Clientset, fip KubefipV1.FloatingIP) ([]byte, error) {
	var err error
	var kubeconfig []byte

	log.Debugf("(getGuestClusterKubeconfig) retrieving guest cluster kubeconfig")

	if fip.ObjectMeta.Annotations["clustername"] == "" {
		errMsg := fmt.Sprintf("(getGuestClusterKubeconfig) clustername annotation not set in fip [%s/%s]",
			fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)
		return kubeconfig, errors.New(errMsg)
	}

	kubeconfigSecretName := fmt.Sprintf("%s-kubeconfig", fip.ObjectMeta.Annotations["clustername"])
	kubeconfigSecretObj, err := clientset.CoreV1().Secrets("fleet-default").Get(context.TODO(), kubeconfigSecretName, metav1.GetOptions{})
	if err != nil {
		return kubeconfig, err
	}

	log.Tracef("(getGuestClusterKubeconfig) got secret object: [%+v]", kubeconfigSecretObj)

	kubeconfig = kubeconfigSecretObj.Data["value"]

	return kubeconfig, err
}

func operateGuestClusters(clientset *kubernetes.Clientset, kubefipConfig *config.KubefipConfigStruct) {
	var kubevipGuestInstallLabel bool

	log.Debugf("(operateGuestClusters) start operating guest clusters")

	metrics.InOperationMode = true

	// create a copy of the AllFips list so we are not run into issues when a fip is removed during cluster operations
	allFipsCopy := kubefip.AllFips

	for i := 0; i < len(allFipsCopy); i++ {
		log.Debugf("(operateGuestClusters) checking fip name [%s] in clusternamespace [%s]",
			allFipsCopy[i].ObjectMeta.Name, allFipsCopy[i].ObjectMeta.Namespace)

		// check if the floatingip object is still a part of the cluster object, otherwise skip the rest
		if err := checkClusterStatus(clientset, allFipsCopy[i]); err != nil {
			log.Errorf("%s", err.Error())
		} else {
			// get the guest cluster kubeconfig
			kubeconfig, err := getGuestClusterKubeconfig(clientset, allFipsCopy[i])
			if err != nil {
				log.Errorf("(operateGuestClusters) error in fetching kubeconfig: %s", err.Error())
			}

			// get all cluster variables
			cluster, err := getClusterVariables(allFipsCopy[i].ObjectMeta.Namespace, clientset)
			if err != nil {
				log.Errorf("(operateGuestClusters) error cannot get cluster object for cluster namespace [%s] to determine clusterlabel: %s",
					allFipsCopy[i].ObjectMeta.Namespace, err.Error())
			}

			// test the connection to the guest cluster
			if err := testGuestClusterConnection(kubeconfig); err != nil {
				// if the error contains "Forbidden" the cluster is in deploy state
				if strings.Contains(err.Error(), "Forbidden") {
					log.Debugf("(operateGuestClusters) guest cluster [%s] is still in deploy state: %s",
						allFipsCopy[i].ObjectMeta.Annotations["clustername"], err.Error())
				} else {
					log.Warningf("(operateGuestClusters) cannot connect to guest cluster [%s]: %s",
						allFipsCopy[i].ObjectMeta.Annotations["clustername"], err.Error())

					metrics.SetGuestClusterStatus(allFipsCopy[i].ObjectMeta.Annotations["clustername"], cluster.HarvesterClusterName,
						metrics.StatusDown)

					metrics.IncrementGuestClusterEventsMetric(allFipsCopy[i].ObjectMeta.Annotations["clustername"],
						cluster.HarvesterClusterName, metrics.EventApiConnection, metrics.StatusError)
				}
			} else {
				metrics.SetGuestClusterStatus(allFipsCopy[i].ObjectMeta.Annotations["clustername"], cluster.HarvesterClusterName,
					metrics.StatusUp)

				metrics.IncrementGuestClusterEventsMetric(allFipsCopy[i].ObjectMeta.Annotations["clustername"],
					cluster.HarvesterClusterName, metrics.EventApiConnection, metrics.StatusSuccess)

				// determine the kube-vip installation type
				kubevipGuestInstallLabel = false
				if kubefipConfig.KubevipGuestInstall == "clusterlabel" {
					// check if the cluster label is set
					if cluster.Labels["kube-vip"] != "" {
						kubeVipLabel, err := strconv.ParseBool(cluster.Labels["kube-vip"])
						if err != nil {
							log.Errorf("(operateGuestClusters) error parsing kube-vip label: %s", err)
						} else {
							kubevipGuestInstallLabel = kubeVipLabel
						}
					}
				}

				log.Debugf("(operateGuestClusters) kubevipGuestInstallLabel: [%+v]", kubevipGuestInstallLabel)

				// try to install kube-vip and the kube-vip-cloud-provider
				if kubefipConfig.KubevipGuestInstall == "enabled" || kubevipGuestInstallLabel {
					if metricUpdate, err := installKubevipInGuestCluster(kubeconfig, kubefipConfig, allFipsCopy[i]); err != nil {
						log.Errorf("(operateGuestClusters) error while managing the kube-vip installation in guest cluster [%s]: %s",
							allFipsCopy[i].ObjectMeta.Annotations["clustername"], err.Error())

						metrics.IncrementGuestClusterEventsMetric(allFipsCopy[i].ObjectMeta.Annotations["clustername"],
							cluster.HarvesterClusterName, metrics.EventKubevipInstall, metrics.StatusError)
					} else {
						if metricUpdate {
							metrics.IncrementGuestClusterEventsMetric(allFipsCopy[i].ObjectMeta.Annotations["clustername"],
								cluster.HarvesterClusterName, metrics.EventKubevipInstall, metrics.StatusSuccess)
						}
					}

					if metricUpdate, err := installKubevipCloudproviderInGuestCluster(kubeconfig, kubefipConfig, allFipsCopy[i]); err != nil {
						log.Errorf("(operateGuestClusters) error while managing the kube-vip-cloud-provider installation in guest cluster [%s]: %s",
							allFipsCopy[i].ObjectMeta.Annotations["clustername"], err.Error())

						metrics.IncrementGuestClusterEventsMetric(allFipsCopy[i].ObjectMeta.Annotations["clustername"],
							cluster.HarvesterClusterName, metrics.EventKubevipCloudproviderInstall, metrics.StatusError)
					} else {
						if metricUpdate {
							metrics.IncrementGuestClusterEventsMetric(allFipsCopy[i].ObjectMeta.Annotations["clustername"],
								cluster.HarvesterClusterName, metrics.EventKubevipCloudproviderInstall, metrics.StatusSuccess)
						}
					}
				}

				// try to manage the kubevip configmap in kube-system
				if metricUpdate, err := createOrUpdateKubevipConfigmapInGuestCluster(kubeconfig, kubefipConfig, allFipsCopy[i]); err != nil {
					log.Errorf("(operateGuestClusters) error while managing the kube-vip config in guest cluster [%s]: %s",
						allFipsCopy[i].ObjectMeta.Annotations["clustername"], err.Error())

					metrics.IncrementGuestClusterEventsMetric(allFipsCopy[i].ObjectMeta.Annotations["clustername"],
						cluster.HarvesterClusterName, metrics.EventConfigmapManagement, metrics.StatusError)
				} else {
					if metricUpdate {
						metrics.IncrementGuestClusterEventsMetric(allFipsCopy[i].ObjectMeta.Annotations["clustername"],
							cluster.HarvesterClusterName, metrics.EventConfigmapManagement, metrics.StatusSuccess)
					}
				}
			}
		}
	}

	log.Debugf("(operateGuestClusters) end operating guest clusters")

	if kubefipConfig.TraceIpamData {
		log.Infof("(IPAM DATA) dumping stored fip and prefix data")
		for i := 0; i < len(allFipsCopy); i++ {
			log.Infof("(IPAM DATA) stored fip name [%s/%s] and ipaddress [%s]",
				allFipsCopy[i].ObjectMeta.Namespace, allFipsCopy[i].ObjectMeta.Name, allFipsCopy[i].Spec.IPAddress)
		}

		for k, v := range kubefip.PrefixList {
			log.Infof("(IPAM DATA) stored prefix/fiprange name [%s] and cidr [%s]", k, v.Cidr)
		}
	}

	metrics.InOperationMode = false
}

func startManageKubevip(clientset *kubernetes.Clientset, kubefipConfig *config.KubefipConfigStruct) {
	log.Infof("(startManageKubevip) start managing the kubevip configs on the guest clusters")

	// this implemention makes sure that the ticker stops and starts again to prevent race conditions
	operateTicker = time.NewTicker(time.Duration(kubefipConfig.OperateGuestClusterInterval) * time.Second)
	quitOperation := make(chan struct{})
	go func() {
		for {
			select {
			case <-operateTicker.C:
				operateGuestClusters(clientset, kubefipConfig)
			case <-quitOperation:
				operateTicker.Stop()
				return
			}
		}
	}()

	metricsCleanupTicker = time.NewTicker(time.Duration(kubefipConfig.OperateGuestClusterInterval/2) * time.Second)
	quitCleanup := make(chan struct{})
	go func() {
		for {
			select {
			case <-metricsCleanupTicker.C:
				metrics.CleanupMetrics()
			case <-quitCleanup:
				metricsCleanupTicker.Stop()
				return
			}
		}
	}()
}

func restartManageKubevip(clientset *kubernetes.Clientset, kubefipConfig *config.KubefipConfigStruct, oldOperateGuestClusterInterval int) {
	log.Infof("(restartManageKubevip) restart managing the kubevip configs on the guest clusters")

	if oldOperateGuestClusterInterval == kubefipConfig.OperateGuestClusterInterval {
		log.Debugf("(restartManageKubevip) oldOperateGuestClusterInterval [%d] matches [%d], no restart necessary",
			oldOperateGuestClusterInterval, kubefipConfig.OperateGuestClusterInterval)

		return
	}

	log.Infof("(restartManageKubevip) stopping the management of the guest clusters")

	// stop the tickers
	operateTicker.Stop()
	metricsCleanupTicker.Stop()

	// start the ticker again
	startManageKubevip(clientset, kubefipConfig)
}

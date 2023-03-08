package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	KubefipV1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	kubefipclientset "github.com/joeyloman/kube-fip-operator/pkg/generated/clientset/versioned"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func checkClusterStatus(k8s_clientset *kubernetes.Clientset, fip KubefipV1.FloatingIP) error {
	var err error

	log.Debugf("(checkClusterStatus) checking if cluster [%s] exists in the clusters.provisioning.cattle.io objects",
		fip.ObjectMeta.Annotations["clustername"])

	cluster, err := k8s_clientset.RESTClient().Get().AbsPath("/apis/provisioning.cattle.io/v1").Namespace("fleet-default").Resource("clusters").Name(fip.ObjectMeta.Annotations["clustername"]).DoRaw(context.TODO())
	if err != nil {
		log.Errorf("(checkClusterStatus) error while fetching cluster objects: %s", err.Error())
	}

	c := ClusterStruct{}
	if err = json.Unmarshal(cluster, &c); err != nil {
		log.Errorf("(checkClusterStatus) error unmarshall json: %s", err.Error())
		return err
	}

	log.Debugf("(checkClusterStatus) cluster object status clustername: [%s] / floatingip namespace: [%s]",
		c.Status.ClusterName, fip.ObjectMeta.Namespace)

	// checking if the namespace in the floatingip objects is the same as in the cluster object
	if c.Status.ClusterName != fip.ObjectMeta.Namespace {
		errMsg := fmt.Sprintf("(checkClusterStatus) error: clustername [%s] from floatingip [%s/%s] cannot be found in the cluster objects",
			fip.ObjectMeta.Annotations["clustername"], fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)
		return errors.New(errMsg)
	}

	return err
}

func getHarvesterClusterName(cloudCredentialSecretName string, k8s_clientset *kubernetes.Clientset) (string, error) {
	var err error
	var harvesterClusterName string

	log.Debugf("(getHarvesterClusterName) fetching the harvester clusterid from the cloudCredentialSecretName: %s", cloudCredentialSecretName)

	cloudCredentialSecret, err := k8s_clientset.CoreV1().Secrets("cattle-global-data").Get(context.TODO(), cloudCredentialSecretName, metav1.GetOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("(getHarvesterNetworkName) error while fetching the cloud credential secret contents: %s", err.Error())
		return harvesterClusterName, errors.New(errMsg)
	}

	harvesterClusterId := string(cloudCredentialSecret.Data["harvestercredentialConfig-clusterId"])

	harvesterCluster, err := k8s_clientset.RESTClient().Get().AbsPath("/apis/management.cattle.io/v3").Resource("clusters").Name(harvesterClusterId).DoRaw(context.TODO())
	if err != nil {
		errMsg := fmt.Sprintf("(getHarvesterNetworkName) error while fetching harvesterCluster object: %s", err.Error())
		return harvesterClusterName, errors.New(errMsg)
	}

	hc := ClusterManagementStruct{}
	if err = json.Unmarshal(harvesterCluster, &hc); err != nil {
		log.Errorf("(getHarvesterNetworkName) error unmarshall json: %s", err.Error())
	}

	harvesterClusterName = hc.Spec.DisplayName

	log.Debugf("(getHarvesterClusterName) found harvesterClusterName: [%s]", harvesterClusterName)

	return harvesterClusterName, err
}

func getHarvesterClusterNameFromFipRange(fip *KubefipV1.FloatingIP, kubefip_clientset *kubefipclientset.Clientset) (string, error) {
	var err error
	var harvesterClusterName string

	fipRange := string(fip.ObjectMeta.Annotations["fiprange"])

	fipRangeObj, err := kubefip_clientset.KubefipV1().FloatingIPRanges().Get(context.TODO(), fipRange, metav1.GetOptions{})
	if err == nil {
		harvesterClusterName = fipRangeObj.ObjectMeta.Annotations["harvesterClusterName"]
	}

	log.Debugf("(getHarvesterClusterNameFromFipRange) fetched harvester clustername [%s] from the fiprange [%s]",
		harvesterClusterName, fipRange)

	return harvesterClusterName, err
}

func getHarvesterNetworkName(machineConfigRefName string, k8s_clientset *kubernetes.Clientset) (string, error) {
	var err error
	var harvesterNetworkName string

	log.Debugf("(getHarvesterNetworkName) checking if there is a network name specified in the matching harvesterconfigs object")

	harvesterConfigs, err := k8s_clientset.RESTClient().Get().AbsPath("/apis/rke-machine-config.cattle.io/v1").Namespace("fleet-default").Resource("harvesterconfigs").DoRaw(context.TODO())
	if err != nil {
		errMsg := fmt.Sprintf("(getHarvesterNetworkName) error while fetching harvesterConfigs objects: %s", err.Error())
		return harvesterNetworkName, errors.New(errMsg)
	}

	h := HarvesterConfigsStruct{}
	if err = json.Unmarshal(harvesterConfigs, &h); err != nil {
		log.Errorf("(getHarvesterNetworkName) error unmarshall json: %s", err.Error())
	}

	for _, item := range h.Items {
		// check if the harvesterconfig object name matches the machineConfigRefName
		if item.Metadata.Name == machineConfigRefName {
			harvesterNetworkNameSplitted := strings.Split(item.NetworkName, "/")

			// get the network name by splitting the Harvester network object <namespace>/<network>
			if len(harvesterNetworkNameSplitted) < 1 {
				log.Errorf("(getHarvesterNetworkName) error harvesterNetworkName format is not correct")
			} else {
				harvesterNetworkName = harvesterNetworkNameSplitted[1]
			}

			break
		}
	}

	return harvesterNetworkName, err
}

func getClusterVariables(nsName string, k8s_clientset *kubernetes.Clientset) (Cluster, error) {
	var err error

	log.Debugf("(getClusterVariables) checking if the new namespace is a new cluster object")

	cluster := Cluster{}

	clusters, err := k8s_clientset.RESTClient().Get().AbsPath("/apis/provisioning.cattle.io/v1").Namespace("fleet-default").Resource("clusters").DoRaw(context.TODO())
	if err != nil {
		errMsg := fmt.Sprintf("(getClusterVariables) error while fetching cluster objects: %s", err.Error())
		return cluster, errors.New(errMsg)
	}

	c := ClustersStruct{}
	if err = json.Unmarshal(clusters, &c); err != nil {
		log.Errorf("(getClusterVariables) error unmarshall json: %s", err.Error())
	}

	for _, item := range c.Items {
		log.Debugf("(getClusterVariables) clusterstruct object: [%+v]", item.Metadata.Name)

		if item.Status.ClusterName == nsName {
			log.Debugf("(getClusterVariables) match found: status clustername [%s] matches namespace [%s]", item.Status.ClusterName, nsName)

			// get the machineConfigRef name so we can lookup the network in HarvesterConfig object
			for _, mps := range item.Spec.RkeConfig.MachinePools {
				log.Debugf("(getClusterVariables) found MachineConfigRef Kind [%s] / Name [%s] ", mps.MachineConfigRef.Kind, mps.MachineConfigRef.Name)

				// TODO: we could also do a doublecheck if the pool has a mps.ControlPlaneRole (now it's based on the first pool hit)
				if mps.MachineConfigRef.Kind == "HarvesterConfig" {
					cluster.MachineConfigRefName = mps.MachineConfigRef.Name

					break
				}
			}

			// store the labels
			cluster.Labels = item.Metadata.Labels

			// check if the cloudCredentialSecretName exists
			if item.Spec.CloudCredentialSecretName == "" {
				log.Debugf("(getClusterVariables) cluster object has no cloudCredentialSecretName in the spec")

				return cluster, err
			} else {
				cloudCredentialSecretNameSplitted := strings.Split(item.Spec.CloudCredentialSecretName, ":")

				// get the cloud credential secret name by splitting the secret object <namespace>:<secret>
				if len(cloudCredentialSecretNameSplitted) < 1 {
					log.Errorf("(getClusterVariables) error cloudCredentialSecretName format is not correct")

					return cluster, err
				} else {
					cluster.CloudCredentialSecretName = cloudCredentialSecretNameSplitted[1]

					// get the actual clustername
					cluster.ClusterName = item.Metadata.Name

					// get the harvester clustername
					harvesterClusterName, err := getHarvesterClusterName(cluster.CloudCredentialSecretName, k8s_clientset)
					if err != nil {
						log.Errorf("%s", err)

						return cluster, err
					}
					cluster.HarvesterClusterName = harvesterClusterName

					return cluster, err
				}
			}
		}
	}

	return cluster, err
}

func checkNewNamespace(ns *corev1.Namespace, kubefip_clientset *kubefipclientset.Clientset, k8s_clientset *kubernetes.Clientset) {
	var fipRangeName string
	var harvesterNetworkName string

	log.Debugf("(checkNewNamespace) checking if the new namespace is a new cluster object")

	// usually it takes some seconds before the harvester objects are created
	time.Sleep(15 * time.Second)

	if strings.HasPrefix(ns.Name, "c-m-") {
		// guest cluster namespace found
		log.Debugf("(checkNewNamespace) new cluster namespace [%s] detected", ns.Name)

		// get the cloud credential name to determine the fiprange
		cluster, err := getClusterVariables(ns.Name, k8s_clientset)
		if err != nil {
			log.Errorf("(checkNewNamespace) cannot get cluster object for cluster namespace [%s]: %s", ns.Name, err.Error())

			return
		}

		// if these objects are empty we have no match
		if cluster.HarvesterClusterName == "" || cluster.ClusterName == "" {
			log.Debugf("(checkNewNamespace) namespace [%s] does not exists as a cluster object", ns.Name)

			return
		}

		log.Debugf("(checkNewNamespace) harvesterClusterName [%s] and cloudCredentialSecretName [%s] and clusterName [%s] and machineConfigRefName [%s] found for namespace [%s]",
			cluster.HarvesterClusterName, cluster.CloudCredentialSecretName, cluster.ClusterName, cluster.MachineConfigRefName, ns.Name)

		// Harvester configuration found, fetching network information
		if cluster.MachineConfigRefName != "" {
			harvesterNetworkName, err = getHarvesterNetworkName(cluster.MachineConfigRefName, k8s_clientset)
			if err != nil {
				log.Errorf("(checkNewNamespace) cannot get harvester network name for cluster namespace [%s]: %s", ns.Name, err.Error())
			}
		}

		log.Debugf("(checkNewNamespace) harvesterNetworkName [%s]", harvesterNetworkName)

		// check if there is already a fip object in the namespace
		fipList, err := kubefip_clientset.KubefipV1().FloatingIPs(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Errorf("(checkNewNamespace) cannot get a list of fips in namespace [%s]: %s", ns.Name, err.Error())

			return
		}

		if len(fipList.Items) > 0 {
			log.Errorf("(checkNewNamespace) namespace [%s] already has %d fip objects registered", ns.Name, len(fipList.Items))

			return
		}

		// get the fipranges and check if the cloud credential has a fiprange, return a fiprange
		fipRangeList, err := kubefip_clientset.KubefipV1().FloatingIPRanges().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Errorf("(checkNewNamespace) cannot get a list of fipranges: %s", err.Error())

			return
		}

		for _, fiprange := range fipRangeList.Items {
			if fiprange.ObjectMeta.Annotations["harvesterClusterName"] == cluster.HarvesterClusterName {
				log.Debugf("(checkNewNamespace) fiprange match found [%s] with harvester cluster name [%s]",
					fiprange.ObjectMeta.Name, fiprange.ObjectMeta.Annotations["harvesterClusterName"])

				// if a harvester network name is found, try to match it with a annotation in the the fiprange
				if harvesterNetworkName != "" {
					if fiprange.ObjectMeta.Annotations["harvesterNetworkName"] == harvesterNetworkName {
						log.Debugf("(checkNewNamespace) fiprange harvester network match found [%s] with harvester network name [%s]",
							fiprange.ObjectMeta.Name, fiprange.ObjectMeta.Annotations["harvesterNetworkName"])

						fipRangeName = fiprange.ObjectMeta.Name

						break
					} else {
						log.Debugf("(checkNewNamespace) no fiprange harvester network match found [%s] with harvester network name [%s]",
							fiprange.ObjectMeta.Name, fiprange.ObjectMeta.Annotations["harvesterNetworkName"])
					}
				} else {
					log.Debugf("(checkNewNamespace) harvesterNetworkName is empty")

					fipRangeName = fiprange.ObjectMeta.Name

					break
				}

				// register the first fiprange hit so we can return it if there is no harvester network match found due a missing annotation
				if fipRangeName == "" {
					fipRangeName = fiprange.ObjectMeta.Name

					log.Debugf("(checkNewNamespace) registered the first fiprange hit: [%s]", fipRangeName)
				}
			}
		}

		if fipRangeName == "" {
			log.Errorf("(checkNewNamespace) no fiprange match found for clustername [%s]", cluster.ClusterName)

			return
		}

		log.Debugf("(checkNewNamespace) clusterName [%s] / fipRangeName [%s]", cluster.ClusterName, fipRangeName)

		// create a new fip object
		fip := KubefipV1.FloatingIP{}
		fip.ObjectMeta.Name = fmt.Sprintf("%s-kubevip", cluster.ClusterName)
		fip.ObjectMeta.Namespace = ns.Name

		annotations := make(map[string]string)
		annotations["clustername"] = cluster.ClusterName
		annotations["fiprange"] = fipRangeName
		annotations["updateConfigMap"] = "true"

		fip.ObjectMeta.Annotations = annotations

		fipCreateObj, err := kubefip_clientset.KubefipV1().FloatingIPs(ns.Name).Create(context.TODO(), &fip, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("(checkNewNamespace) error creating fip [%s/%s]: %s", fip.ObjectMeta.Namespace, fip.ObjectMeta.Name, err.Error())
			return
		}

		log.Infof("(checkNewNamespace) successfully created new fip object [%s/%s] for cluster [%s]",
			fipCreateObj.ObjectMeta.Namespace, fipCreateObj.ObjectMeta.Name, cluster.ClusterName)
	} else {
		log.Debugf("(checkNewNamespace) new namespace [%s] is not a guest cluster", ns.Name)
	}
}

package config

import (
	"context"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubefipConfigStruct struct {
	LogLevel                        string `json:"LogLevel"`
	TraceIpamData                   bool   `json:"TraceIpamData"`
	OperateGuestClusterInterval     int    `json:"OperateGuestClusterInterval"`
	KubevipGuestInstall             string `json:"KubevipGuestInstall"`
	KubevipNamespace                string `json:"KubevipNamespace"`
	KubevipChartRepoUrl             string `json:"KubevipChartRepoUrl"`
	KubevipChartValues              string `json:"KubevipChartValues"`
	KubevipCloudProviderChartValues string `json:"KubevipCloudProviderChartValues"`
	KubevipUpdate                   bool   `json:"KubevipUpdate"`
	RemoveHarvesterCloudProvider    bool   `json:"RemoveHarvesterCloudProvider"`
	HarvesterCloudProviderNamespace string `json:"HarvesterCloudProviderNamespace"`
}

func GetKubefipConfigmap(clientset *kubernetes.Clientset) (*corev1.ConfigMap, error) {
	var err error

	kubefipConfigMap, err := clientset.CoreV1().ConfigMaps("kube-fip").Get(context.TODO(), "kube-fip-config", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return kubefipConfigMap, err
}

func ParseKubfipConfigMap(kubefipConfigmap *corev1.ConfigMap) KubefipConfigStruct {
	log.Infof("(ParseKubfipConfigMap) kube-fip-config ConfigMap changed, parsing new config ..")

	kubefipConfig := KubefipConfigStruct{}

	// set the defaults
	kubefipConfig.LogLevel = "Info"
	kubefipConfig.TraceIpamData = false
	kubefipConfig.OperateGuestClusterInterval = 480
	kubefipConfig.KubevipGuestInstall = "clusterlabel" // can be disabled (don't install), enabled (install on every cluster) or clusterlabel (checks for kube-vip=true label)
	kubefipConfig.KubevipNamespace = "kube-system"
	kubefipConfig.KubevipChartRepoUrl = "https://kube-vip.io/helm-charts"
	kubefipConfig.KubevipChartValues = "{\"image\":{\"repository\":\"plndr/kube-vip\",\"tag\":\"v0.3.7\"},\"config\":{\"vip_interface\":\"enp1s0\"},\"nodeSelector\":{\"node-role.kubernetes.io/master\":\"true\"}}"
	kubefipConfig.KubevipCloudProviderChartValues = "{\"image\":{\"repository\":\"kubevip/kube-vip-cloud-provider\",\"tag\":\"0.1\"}}"
	kubefipConfig.KubevipUpdate = false
	kubefipConfig.RemoveHarvesterCloudProvider = false
	kubefipConfig.HarvesterCloudProviderNamespace = "kube-system"

	if kubefipConfigmap == nil {
		log.Debugf("(ParseKubfipConfigMap) config options: LogLevel [%s] / TraceIpamData [%+v] / OperateGuestClusterInterval [%d] / "+
			"KubevipGuestInstall [%s] / KubevipNamespace [%s] / KubevipChartRepoUrl [%s] / KubevipChartValues [%s] / "+
			"KubevipCloudProviderChartValues [%s] / KubevipUpdate [%+v] / RemoveHarvesterCloudProvider [%+v] / "+
			"HarvesterCloudProviderNamespace [%+v]",
			kubefipConfig.LogLevel, kubefipConfig.TraceIpamData, kubefipConfig.OperateGuestClusterInterval, kubefipConfig.KubevipGuestInstall,
			kubefipConfig.KubevipNamespace, kubefipConfig.KubevipChartRepoUrl, kubefipConfig.KubevipChartValues,
			kubefipConfig.KubevipCloudProviderChartValues, kubefipConfig.KubevipUpdate, kubefipConfig.RemoveHarvesterCloudProvider,
			kubefipConfig.HarvesterCloudProviderNamespace)

		return kubefipConfig
	}

	if kubefipConfigmap.Data["logLevel"] != "" {
		kubefipConfig.LogLevel = kubefipConfigmap.Data["logLevel"]
	}

	if kubefipConfigmap.Data["traceIpamData"] != "" {
		traceIpamData, err := strconv.ParseBool(kubefipConfigmap.Data["traceIpamData"])
		if err != nil {
			log.Errorf("(parseKubfipConfigMap) error parsing traceIpamData: %s", err)
		}

		kubefipConfig.TraceIpamData = traceIpamData
	}

	if kubefipConfigmap.Data["operateGuestClusterInterval"] != "" {
		operateGuestClusterInterval, err := strconv.Atoi(kubefipConfigmap.Data["operateGuestClusterInterval"])
		if err != nil {
			log.Errorf("(parseKubfipConfigMap) error parsing operateGuestClusterInterval: %s", err)
		}

		kubefipConfig.OperateGuestClusterInterval = operateGuestClusterInterval
	}

	if kubefipConfigmap.Data["kubevipGuestInstall"] != "" {
		kubefipConfig.KubevipGuestInstall = strings.ToLower(kubefipConfigmap.Data["kubevipGuestInstall"])
	}

	if kubefipConfigmap.Data["kubevipNamespace"] != "" {
		kubefipConfig.KubevipNamespace = kubefipConfigmap.Data["kubevipNamespace"]
	}

	if kubefipConfigmap.Data["kubevipChartRepoUrl"] != "" {
		kubefipConfig.KubevipChartRepoUrl = kubefipConfigmap.Data["kubevipChartRepoUrl"]
	}

	if kubefipConfigmap.Data["kubevipChartValues"] != "" {
		kubefipConfig.KubevipChartValues = kubefipConfigmap.Data["kubevipChartValues"]
	}

	if kubefipConfigmap.Data["kubevipCloudProviderChartValues"] != "" {
		kubefipConfig.KubevipCloudProviderChartValues = kubefipConfigmap.Data["kubevipCloudProviderChartValues"]
	}

	if kubefipConfigmap.Data["kubevipUpdate"] != "" {
		kubevipUpdate, err := strconv.ParseBool(kubefipConfigmap.Data["kubevipUpdate"])
		if err != nil {
			log.Errorf("(parseKubfipConfigMap) error parsing kubevipUpdate: %s", err)
		}

		kubefipConfig.KubevipUpdate = kubevipUpdate
	}

	if kubefipConfigmap.Data["removeHarvesterCloudProvider"] != "" {
		removeHarvesterCloudProvider, err := strconv.ParseBool(kubefipConfigmap.Data["removeHarvesterCloudProvider"])
		if err != nil {
			log.Errorf("(parseKubfipConfigMap) error parsing removeHarvesterCloudProvider: %s", err)
		}

		kubefipConfig.RemoveHarvesterCloudProvider = removeHarvesterCloudProvider
	}

	if kubefipConfigmap.Data["harvesterCloudProviderNamespace"] != "" {
		kubefipConfig.HarvesterCloudProviderNamespace = kubefipConfigmap.Data["harvesterCloudProviderNamespace"]
	}

	log.Debugf("(ParseKubfipConfigMap) config options: LogLevel [%s] / TraceIpamData [%+v] / OperateGuestClusterInterval [%d] / "+
		"KubevipGuestInstall [%s] / KubevipNamespace [%s] / KubevipChartRepoUrl [%s] / KubevipChartValues [%s] / "+
		"KubevipCloudProviderChartValues [%s] / KubevipUpdate [%+v] / RemoveHarvesterCloudProvider [%+v] / "+
		"HarvesterCloudProviderNamespace [%+v]",
		kubefipConfig.LogLevel, kubefipConfig.TraceIpamData, kubefipConfig.OperateGuestClusterInterval, kubefipConfig.KubevipGuestInstall,
		kubefipConfig.KubevipNamespace, kubefipConfig.KubevipChartRepoUrl, kubefipConfig.KubevipChartValues,
		kubefipConfig.KubevipCloudProviderChartValues, kubefipConfig.KubevipUpdate, kubefipConfig.RemoveHarvesterCloudProvider,
		kubefipConfig.HarvesterCloudProviderNamespace)

	return kubefipConfig
}

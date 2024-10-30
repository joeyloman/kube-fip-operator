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
	LogLevel                         string `json:"LogLevel"`
	TraceIpamData                    bool   `json:"TraceIpamData"`
	OperateGuestClusterInterval      int    `json:"OperateGuestClusterInterval"`
	MetricsPort                      int    `json:"MetricsPort"`
	KubevipGuestInstall              string `json:"KubevipGuestInstall"`
	KubevipNamespace                 string `json:"KubevipNamespace"`
	KubevipReleaseName               string `json:"KubevipReleaseName"`
	KubevipChartRepoUrl              string `json:"KubevipChartRepoUrl"`
	KubevipChartRef                  string `json:"KubevipChartRef"`
	KubevipChartVersion              string `json:"KubevipChartVersion"`
	KubevipChartValues               string `json:"KubevipChartValues"`
	KubevipCloudProviderReleaseName  string `json:"KubevipCloudProviderReleaseName"`
	KubevipCloudProviderChartRef     string `json:"KubevipCloudProviderChartRef"`
	KubevipCloudProviderChartVersion string `json:"KubevipCloudProviderChartVersion"`
	KubevipCloudProviderChartValues  string `json:"KubevipCloudProviderChartValues"`
	KubevipUpdate                    bool   `json:"KubevipUpdate"`
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
	kubefipConfig.MetricsPort = 8080
	kubefipConfig.KubevipGuestInstall = "clusterlabel" // can be disabled (don't install), enabled (install on every cluster) or clusterlabel (checks for kube-vip=true label)
	kubefipConfig.KubevipNamespace = "kube-system"
	kubefipConfig.KubevipReleaseName = "kube-vip"
	kubefipConfig.KubevipChartRepoUrl = "https://kube-vip.io/helm-charts"
	kubefipConfig.KubevipChartRef = ""
	kubefipConfig.KubevipChartVersion = ""
	kubefipConfig.KubevipChartValues = "{\"image\":{\"repository\":\"plndr/kube-vip\",\"tag\":\"v0.6.4\"},\"config\":{\"vip_interface\":\"enp1s0\"},\"nodeSelector\":{\"node-role.kubernetes.io/master\":\"true\"}}"
	kubefipConfig.KubevipCloudProviderReleaseName = "kube-vip-cloud-provider"
	kubefipConfig.KubevipCloudProviderChartRef = ""
	kubefipConfig.KubevipCloudProviderChartVersion = ""
	kubefipConfig.KubevipCloudProviderChartValues = "{\"image\":{\"repository\":\"kubevip/kube-vip-cloud-provider\",\"tag\":\"v0.0.7\"}}"
	kubefipConfig.KubevipUpdate = false

	if kubefipConfigmap == nil {
		log.Debugf("(ParseKubfipConfigMap) config options: LogLevel [%s] / TraceIpamData [%+v] / OperateGuestClusterInterval [%d] / "+
			"MetricsPort [%d] / KubevipGuestInstall [%s] / KubevipNamespace [%s] / KubevipReleaseName [%s] / KubevipChartRepoUrl [%s] / "+
			"KubevipChartRef [%s] / KubevipChartVersion [%s] / KubevipChartValues [%s] / KubevipCloudProviderReleaseName [%s] / "+
			"KubevipCloudProviderChartRef [%s] / KubevipCloudProviderChartVersion [%s] / KubevipCloudProviderChartValues [%s] / KubevipUpdate [%+v]",
			kubefipConfig.LogLevel, kubefipConfig.TraceIpamData, kubefipConfig.OperateGuestClusterInterval, kubefipConfig.MetricsPort,
			kubefipConfig.KubevipGuestInstall, kubefipConfig.KubevipNamespace, kubefipConfig.KubevipReleaseName, kubefipConfig.KubevipChartRepoUrl,
			kubefipConfig.KubevipChartRef, kubefipConfig.KubevipChartVersion, kubefipConfig.KubevipChartValues, kubefipConfig.KubevipCloudProviderReleaseName,
			kubefipConfig.KubevipCloudProviderChartRef, kubefipConfig.KubevipCloudProviderChartVersion, kubefipConfig.KubevipCloudProviderChartValues,
			kubefipConfig.KubevipUpdate)

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

	if kubefipConfigmap.Data["metricsPort"] != "" {
		metricsPort, err := strconv.Atoi(kubefipConfigmap.Data["metricsPort"])
		if err != nil {
			log.Errorf("(parseKubfipConfigMap) error parsing metricsPort: %s", err)
		}

		kubefipConfig.MetricsPort = metricsPort
	}

	if kubefipConfigmap.Data["kubevipGuestInstall"] != "" {
		kubefipConfig.KubevipGuestInstall = strings.ToLower(kubefipConfigmap.Data["kubevipGuestInstall"])
	}

	if kubefipConfigmap.Data["kubevipNamespace"] != "" {
		kubefipConfig.KubevipNamespace = kubefipConfigmap.Data["kubevipNamespace"]
	}

	if kubefipConfigmap.Data["kubevipReleaseName"] != "" {
		kubefipConfig.KubevipReleaseName = kubefipConfigmap.Data["kubevipReleaseName"]
	}

	if kubefipConfigmap.Data["kubevipChartRepoUrl"] != "" {
		kubefipConfig.KubevipChartRepoUrl = kubefipConfigmap.Data["kubevipChartRepoUrl"]
	}

	if kubefipConfigmap.Data["kubevipChartRef"] != "" {
		kubefipConfig.KubevipChartRef = kubefipConfigmap.Data["kubevipChartRef"]
	}

	if kubefipConfigmap.Data["kubevipChartVersion"] != "" {
		kubefipConfig.KubevipChartVersion = kubefipConfigmap.Data["kubevipChartVersion"]
	}

	if kubefipConfigmap.Data["kubevipChartValues"] != "" {
		kubefipConfig.KubevipChartValues = kubefipConfigmap.Data["kubevipChartValues"]
	}

	if kubefipConfigmap.Data["kubevipCloudProviderReleaseName"] != "" {
		kubefipConfig.KubevipCloudProviderReleaseName = kubefipConfigmap.Data["kubevipCloudProviderReleaseName"]
	}

	if kubefipConfigmap.Data["kubevipCloudProviderChartRef"] != "" {
		kubefipConfig.KubevipCloudProviderChartRef = kubefipConfigmap.Data["kubevipCloudProviderChartRef"]
	}

	if kubefipConfigmap.Data["kubevipCloudProviderChartVersion"] != "" {
		kubefipConfig.KubevipCloudProviderChartVersion = kubefipConfigmap.Data["kubevipCloudProviderChartVersion"]
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

	log.Debugf("(ParseKubfipConfigMap) config options: LogLevel [%s] / TraceIpamData [%+v] / OperateGuestClusterInterval [%d] / "+
		"MetricsPort [%d] / KubevipGuestInstall [%s] / KubevipNamespace [%s] / KubevipReleaseName [%s] / KubevipChartRepoUrl [%s] / "+
		"KubevipChartRef [%s] / KubevipChartVersion [%s] / KubevipChartValues [%s] / KubevipCloudProviderReleaseName [%s] / "+
		"KubevipCloudProviderChartRef [%s] / KubevipCloudProviderChartVersion [%s] / KubevipCloudProviderChartValues [%s] / KubevipUpdate [%+v]",
		kubefipConfig.LogLevel, kubefipConfig.TraceIpamData, kubefipConfig.OperateGuestClusterInterval, kubefipConfig.MetricsPort,
		kubefipConfig.KubevipGuestInstall, kubefipConfig.KubevipNamespace, kubefipConfig.KubevipReleaseName, kubefipConfig.KubevipChartRepoUrl,
		kubefipConfig.KubevipChartRef, kubefipConfig.KubevipChartVersion, kubefipConfig.KubevipChartValues, kubefipConfig.KubevipCloudProviderReleaseName,
		kubefipConfig.KubevipCloudProviderChartRef, kubefipConfig.KubevipCloudProviderChartVersion, kubefipConfig.KubevipCloudProviderChartValues,
		kubefipConfig.KubevipUpdate)

	return kubefipConfig
}

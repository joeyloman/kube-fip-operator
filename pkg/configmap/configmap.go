package configmap

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	KubefipV1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
)

func NewKubevipConfigmap(fip *KubefipV1.FloatingIP, kubevipConfigMapName string, kubevipConfigMapNamespace string) corev1.ConfigMap {
	log.Debugf("(generateKubevipConfigmap) generating new kubevip configmap")

	// generate the data objects
	configMapData := make(map[string]string)
	cidr := fmt.Sprintf("%s/32", fip.Spec.IPAddress)
	configMapData["cidr-global"] = cidr

	// create the corev1.ConfigMap type
	kubevipConfigMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubevipConfigMapName,
			Namespace: kubevipConfigMapNamespace,
		},
		Data: configMapData,
	}

	log.Tracef("(generateKubevipConfigmap) generated configmap [%+v]", kubevipConfigMap)

	return kubevipConfigMap
}

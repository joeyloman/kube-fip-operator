package kubefip

import (
	"context"
	"errors"
	"fmt"

	KubefipV1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	kubefipclientset "github.com/joeyloman/kube-fip-operator/pkg/generated/clientset/versioned"
	"github.com/joeyloman/kube-fip-operator/pkg/metrics"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AllocateFip(fip *KubefipV1.FloatingIP, clientset *kubefipclientset.Clientset) error {
	var err error

	log.Tracef("(AllocateFip) fipobj added: [%+v]", fip)

	// get the clustername from the fip object annotation and check if it exists
	cName := fip.ObjectMeta.Annotations["clustername"]
	if cName == "" {
		errMsg := fmt.Sprintf("clustername not found in annotations for [%s/%s]",
			fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)
		return errors.New(errMsg)
	}

	// get the fiprange from the fip object annotation
	frName := fip.ObjectMeta.Annotations["fiprange"]
	if frName == "" {
		errMsg := fmt.Sprintf("fiprange not found in annotations for [%s/%s]",
			fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)
		return errors.New(errMsg)
	}

	// check if the fiprange exists
	_, err = GetFipRange(frName)
	if err != nil {
		return err
	}

	// check if the spec has an IPAddress specified
	if fip.Spec.IPAddress == "" {
		ip, err := IPAM.GetIP(frName, "")
		if err != nil {
			log.Errorf("(AllocateFip) cannot acquire new ip address for [%s/%s]",
				fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)
			return err
		} else {
			log.Infof("(AllocateFip) successfully allocated fip [%s/%s] with new IP address: %s",
				fip.ObjectMeta.Namespace, fip.ObjectMeta.Name, ip)
		}

		// update the fip object in kubernetes
		fip.Spec.IPAddress = ip
		updatedFip, err := clientset.KubefipV1().FloatingIPs(fip.ObjectMeta.Namespace).Update(context.TODO(), fip, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		log.Infof("(AllocateFip) successfully updated Kubernetes fip object [%s/%s] with IP address [%s]",
			updatedFip.ObjectMeta.Namespace, updatedFip.ObjectMeta.Name, updatedFip.Spec.IPAddress)

		// update the metrics
		fipRange, err := GetFipRange(frName)
		if err != nil {
			log.Errorf("(AllocateFip) could not increment Fipranges metrics: %s", err)
		} else {
			metrics.IncrementFiprangesReserved(frName, fipRange.Spec.IPRange, fipRange.ObjectMeta.Annotations["harvesterClusterName"],
				fipRange.ObjectMeta.Annotations["harvesterNetworkName"])
		}

		// add/update the fip in the allFips list
		if err := UpdateAllFips(updatedFip); err != nil {
			return err
		}
	} else {
		// register the allocated fip in the prefix
		ip, err := IPAM.GetIP(frName, fip.Spec.IPAddress)
		if err != nil {
			log.Errorf("(AllocateFip) cannot acquire existing IP address [%s] for [%s/%s]",
				fip.Spec.IPAddress, fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)
			return err
		} else {
			log.Infof("(AllocateFip) successfully allocated fip [%s/%s] with new IP address: %s",
				fip.ObjectMeta.Namespace, fip.ObjectMeta.Name, ip)
		}

		// update the metrics
		fipRange, err := GetFipRange(frName)
		if err != nil {
			log.Errorf("(AllocateFip) could not increment Fipranges metrics: %s", err)
		} else {
			metrics.IncrementFiprangesReserved(frName, fipRange.Spec.IPRange, fipRange.ObjectMeta.Annotations["harvesterClusterName"],
				fipRange.ObjectMeta.Annotations["harvesterNetworkName"])
		}

		// add the fip in the allFips list
		if err := UpdateAllFips(fip); err != nil {
			return err
		}
	}

	return err
}

func RemoveFip(fip *KubefipV1.FloatingIP) error {
	var err error

	log.Tracef("(RemoveFip) fipobj removed: [%+v]", fip)

	// get the fiprange from the fip object annotation
	frName := fip.ObjectMeta.Annotations["fiprange"]
	if frName == "" {
		return errors.New("fiprange not found in annotations")
	}

	// check if the fiprange exists
	_, err = GetFipRange(frName)
	if err != nil {
		log.Errorf("%s", err.Error())
	} else {
		if err := IPAM.ReleaseIP(frName, fip.Spec.IPAddress); err != nil {
			log.Errorf("(RemoveFip) error while removing fip [%s] with ip [%s] from subnet [%s]: %s",
				fip.ObjectMeta.Name, fip.Spec.IPAddress, frName, err.Error())
		} else {
			log.Infof("(RemoveFip) successfully removed fip [%s] with ip [%s] from pfx cidr [%s]",
				fip.ObjectMeta.Name, fip.Spec.IPAddress, frName)

			// update the metrics
			fipRange, err := GetFipRange(frName)
			if err != nil {
				log.Errorf("(RemoveFip) could not decrement Fipranges metrics: %s", err)
			} else {
				metrics.DecrementFiprangesReserved(frName, fipRange.Spec.IPRange, fipRange.ObjectMeta.Annotations["harvesterClusterName"],
					fipRange.ObjectMeta.Annotations["harvesterNetworkName"])
			}
		}
	}

	if err := RemoveFipFromAllFips(fip); err != nil {
		return err
	}

	return err
}

func UpdateFip(oldFip *KubefipV1.FloatingIP, newFip *KubefipV1.FloatingIP, clientset *kubefipclientset.Clientset) error {
	var err error

	log.Tracef("(UpdateFip) fipobj removed: oldFip [%+v] / newFip [%+v]", oldFip, newFip)

	// remove the FIP
	if err := RemoveFip(oldFip); err != nil {
		log.Errorf("(updateFip) Error removing fip: %s", err.Error())
	}

	// allocate the new FIP
	if err := AllocateFip(newFip, clientset); err != nil {
		log.Errorf("(updateFip) Error allocating fip: %s", err.Error())
	}

	return err
}

package kubefip

import (
	"context"
	"errors"
	"fmt"

	KubefipV1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	kubefipclientset "github.com/joeyloman/kube-fip-operator/pkg/generated/clientset/versioned"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AllocateFip(fip *KubefipV1.FloatingIP, clientset *kubefipclientset.Clientset) error {
	var err error

	log.Tracef("(AllocateFip) fipobj added: [%+v]\n", fip)

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

	// lookup the fiprange name to get the cidr and check if it exists
	pfx := PrefixList[frName]
	if pfx.Cidr == "" {
		errMsg := fmt.Sprintf("fiprange [%s] not found in prefix list for [%s/%s]",
			frName, fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)
		return errors.New(errMsg)
	}

	// check if the spec has an IPAddress specified
	if fip.Spec.IPAddress == "" {
		// allocate a new FIP in the prefix
		ip, err := ipam.AcquireIP(pfx.Cidr)
		if err != nil {
			log.Errorf("(AllocateFip) cannot acquire new ip address for [%s/%s]",
				fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)
			return err
		} else {
			log.Infof("(AllocateFip) successfully allocated fip [%s/%s] with new IP address: %s",
				fip.ObjectMeta.Namespace, fip.ObjectMeta.Name, ip.IP)
		}

		// update the fip object in kubernetes
		fip.Spec.IPAddress = ip.IP.String()
		updatedFip, err := clientset.KubefipV1().FloatingIPs(fip.ObjectMeta.Namespace).Update(context.TODO(), fip, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		log.Infof("(AllocateFip) successfully updated Kubernetes fip object [%s/%s] with IP address [%s]",
			updatedFip.ObjectMeta.Namespace, updatedFip.ObjectMeta.Name, updatedFip.Spec.IPAddress)

		// add/update the fip in the allFips list
		if err := UpdateAllFips(updatedFip); err != nil {
			return err
		}
	} else {
		// register the allocated fip in the prefix
		ip, err := ipam.AcquireSpecificIP(pfx.Cidr, fip.Spec.IPAddress)
		if err != nil {
			return err
		} else {
			log.Infof("(AllocateFip) successfully allocated fip [%s/%s] with existing IP address: %s",
				fip.ObjectMeta.Namespace, fip.ObjectMeta.Name, ip.IP)
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

	log.Tracef("(RemoveFip) fipobj removed: [%+v]\n", fip)

	// get the fiprange from the fip object annotation
	frName := fip.ObjectMeta.Annotations["fiprange"]
	if frName == "" {
		return errors.New("fiprange not found in annotations")
	}

	// lookup the fiprange name to get the cidr and check if it exists
	pfx := PrefixList[frName]
	if pfx.Cidr != "" {
		if err := ipam.ReleaseIPFromPrefix(pfx.Cidr, fip.Spec.IPAddress); err != nil {
			log.Errorf("(RemoveFip) error while removing fip [%s/%s] from pfx cidr [%s]: %s",
				fip.ObjectMeta.Name, fip.Spec.IPAddress, frName, err.Error())
		} else {
			log.Infof("(RemoveFip) successfully removed fip [%s/%s] from pfx cidr [%s]",
				fip.ObjectMeta.Name, fip.Spec.IPAddress, frName)
		}
	} else {
		log.Errorf("(RemoveFip) iprange not found in prefix list")
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

package kubefip

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"k8s.io/client-go/kubernetes"

	KubefipV1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	kubefipclientset "github.com/joeyloman/kube-fip-operator/pkg/generated/clientset/versioned"
	goipam "github.com/metal-stack/go-ipam"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	AllFipRanges []KubefipV1.FloatingIPRange // note: this list is only needed during startup, lookups are replaced by the prefixList map
	AllFips      []KubefipV1.FloatingIP
	ipam         goipam.Ipamer
	PrefixList   map[string]goipam.Prefix
)

func GatherAllFipRanges(clientset *kubefipclientset.Clientset) error {
	var err error

	log.Infof("(GatherAllFipRanges) gathering and storing al floatingipranges..")

	fipRangeList, err := clientset.KubefipV1().FloatingIPRanges().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, fiprange := range fipRangeList.Items {
		log.Infof("(GatherAllFipRanges) fiprange found: %s", fiprange.Name)
		log.Tracef("(GatherAllFipRanges) fiprange object: %+v", fiprange)

		AllFipRanges = append(AllFipRanges, fiprange)
	}

	return err
}

func GatherAllFips(k8s_clientset *kubernetes.Clientset, kubefip_clientset *kubefipclientset.Clientset) error {
	var err error

	log.Infof("(GatherAllFips) gathering and storing al floatingips..")

	// TODO namespaces are also labeled, so maybe do a label selection here?
	// get all namespaces and check in the c-m-* names if there are floatingip objects
	nsList, err := k8s_clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, ns := range nsList.Items {
		log.Tracef("(GatherAllFips) namespace found: %s", ns.Name)
		if strings.HasPrefix(ns.Name, "c-m-") {
			// guest cluster namespace found
			fipList, err := kubefip_clientset.KubefipV1().FloatingIPs(ns.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			for _, fip := range fipList.Items {
				log.Infof("(GatherAllFips) fip [%s] found in namespace [%s]", fip.Name, ns.Name)
				log.Tracef("(GatherAllFips) fip object: %+v", fip)

				AllFips = append(AllFips, fip)
			}
		}
	}

	return err
}

func UpdateAllFips(fip *KubefipV1.FloatingIP) error {
	var err error
	var updatedFipFound bool = false

	log.Debugf("(UpdateAllFips) updating fip [%s/%s] in allFips list", fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)

	var newAllFips []KubefipV1.FloatingIP

	for i := 0; i < len(AllFips); i++ {
		// if the updated fip matches the one in the list, add the new fip to the new list
		if fip.ObjectMeta.Namespace == AllFips[i].ObjectMeta.Namespace && fip.ObjectMeta.Name == AllFips[i].ObjectMeta.Name {
			// if the updated fip matches the one in the list, add the new fip to the new list
			log.Debugf("(UpdateAllFips) fip to update found, adding new fip to the list")

			newAllFips = append(newAllFips, *fip)

			updatedFipFound = true
		} else {
			// if there is no match, add the fip to the new list
			log.Debugf("(UpdateAllFips) adding existing fip to the list")

			newAllFips = append(newAllFips, AllFips[i])
		}
	}

	// if no updated fip is found then it should be a new one
	if !updatedFipFound {
		log.Debugf("(UpdateAllFips) adding new fip to the list")

		newAllFips = append(newAllFips, *fip)
	}

	// all good, assign the new list
	AllFips = newAllFips

	return err
}

func RemoveFipFromAllFips(fip *KubefipV1.FloatingIP) error {
	var err error
	var FipFound bool = false

	log.Debugf("(RemoveFipFromAllFips) removing fip [%s/%s] from allFips list", fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)

	var newAllFips []KubefipV1.FloatingIP

	for i := 0; i < len(AllFips); i++ {
		// if the fip matches the one in the list, skip it
		if fip.ObjectMeta.Namespace == AllFips[i].ObjectMeta.Namespace && fip.ObjectMeta.Name == AllFips[i].ObjectMeta.Name {
			// if the updated fip matches the one in the list, add the new fip to the new list
			log.Debugf("(RemoveFipFromAllFips) fip to remove found, skip appending fip to the list")

			FipFound = true
		} else {
			// if there is no match, add the fip to the new list
			log.Debugf("(RemoveFipFromAllFips) adding existing fip [%s/%s] to the list", AllFips[i].ObjectMeta.Namespace, AllFips[i].ObjectMeta.Name)

			newAllFips = append(newAllFips, AllFips[i])
		}
	}

	// if no fip is found then we should return a error
	if !FipFound {
		// should not be reached!
		errMsg := fmt.Sprintf("fip [%s/%s] not found in the allFips list", fip.ObjectMeta.Namespace, fip.ObjectMeta.Name)
		return errors.New(errMsg)
	}

	log.Infof("(RemoveFipFromAllFips) successfully removed fip [%s/%s] from allFips list", fip.ObjectMeta.Name, fip.Spec.IPAddress)

	// all good, assign the new list
	AllFips = newAllFips

	return err
}

func InitIpam() {
	// create a ipamer with in memory storage
	ipam = goipam.New()
}

func CreateIpamPrefixesFromFipRanges() {
	log.Debugf("(CreateIpamPrefixesFromFipRanges) start creating ipam prefixes from fipranges..")

	// initialize the list with prefixes
	PrefixList = make(map[string]goipam.Prefix)

	for i := 0; i < len(AllFipRanges); i++ {
		log.Tracef("(CreateIpamPrefixesFromFipRanges) fiprange obj: [%+v]", AllFipRanges[i])

		if err := AllocateFipRange(&AllFipRanges[i]); err != nil {
			log.Errorf("(watchFipRangeEvents) error allocating fip: %s", err.Error())
		}
	}
}

func StoreAllocatedIpsInIpamPrefixes(clientset *kubefipclientset.Clientset) {
	log.Debugf("(StoreAllocatedIpsInIpamPrefixes) start storing fips in ipam prefixes..")

	for i := 0; i < len(AllFips); i++ {
		log.Tracef("(StoreAllocatedIpsInIpamPrefixes) fip obj: [%+v]", AllFips[i])

		if err := AllocateFip(&AllFips[i], clientset); err != nil {
			log.Errorf("(StoreAllocatedIpsInIpamPrefixes) error allocating fip: %s", err.Error())
		}
	}
}

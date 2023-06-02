package kubefip

import (
	"errors"
	"fmt"

	KubefipV1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	"github.com/joeyloman/kube-fip-operator/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

func GetFipRange(fipRangeName string) (KubefipV1.FloatingIPRange, error) {
	log.Debugf("(GetFipRange) retrieving fipRangeName: [%s]", fipRangeName)

	for i := 0; i < len(AllFipRanges); i++ {
		// check if the fiprange has a match
		if fipRangeName == AllFipRanges[i].ObjectMeta.Name {
			log.Debugf("(GetFipRange) fiprange match found, returning object")

			return AllFipRanges[i], nil
		}
	}

	errMsg := fmt.Sprintf("(GetFipRange) fiprange [%s] not found!", fipRangeName)

	return KubefipV1.FloatingIPRange{}, errors.New(errMsg)
}

func AllocateFipRange(fipRange *KubefipV1.FloatingIPRange) error {
	var err error

	log.Tracef("(AllocateFipRange) fiprangeobj added: [%+v]", fipRange)

	// get the fiprange from the fiprange object
	if fipRange.Spec.IPRange == "" {
		return errors.New("fiprange not found in spec")
	}

	// create a new prefix and add it to the prefixList
	prefix, err := ipam.NewPrefix(ctx, fipRange.Spec.IPRange)
	if err != nil {
		log.Errorf("error creating ipam prefix: %s", err.Error())
		return err
	} else {
		PrefixList[fipRange.ObjectMeta.Name] = *prefix
	}

	log.Infof("(AllocateFipRange) successfully allocated fiprange [%s] with cidr [%s]",
		fipRange.ObjectMeta.Name, prefix.Cidr)

	metrics.SetFiprangesCapacity(fipRange.ObjectMeta.Name, fipRange.Spec.IPRange, fipRange.ObjectMeta.Annotations["harvesterClusterName"],
		fipRange.ObjectMeta.Annotations["harvesterNetworkName"])

	// add/update the fiprange in the allFipRanges list
	if err := UpdateAllFipRanges(fipRange); err != nil {
		return err
	}

	return err
}

func RemoveFipRange(fipRange *KubefipV1.FloatingIPRange) error {
	var err error

	log.Tracef("(RemoveFipRange) fiprangeobj removed: [%+v]", fipRange)

	// get the fiprange from the fiprange object
	fr_name := fipRange.Spec.IPRange
	if fr_name == "" {
		return errors.New("fiprange not found in spec")
	}

	// delete the prefix from the IPAM object
	prefix, err := ipam.DeletePrefix(ctx, fr_name)
	if err != nil {
		return err
	}

	log.Infof("(RemoveFipRange) successfully removed fiprange [%s] with cidr [%s]",
		fipRange.ObjectMeta.Name, prefix)

	metrics.RemoveFiprangeMetrics(fipRange.ObjectMeta.Name, fipRange.Spec.IPRange, fipRange.ObjectMeta.Annotations["harvesterClusterName"],
		fipRange.ObjectMeta.Annotations["harvesterNetworkName"])

	if err := RemoveFipRangeFromAllFipRanges(fipRange); err != nil {
		return err
	}

	return err
}

func UpdateFipRange(oldFipRange *KubefipV1.FloatingIPRange, newFipRange *KubefipV1.FloatingIPRange) error {
	var err error

	log.Tracef("(UpdateFipRange) fiprangeobj removed: oldFipRange [%+v] / newFipRange [%+v]",
		oldFipRange, newFipRange)

	// remove the FIP
	if err := RemoveFipRange(oldFipRange); err != nil {
		log.Errorf("(UpdateFipRange) Error removing oldFipRange: %s", err.Error())
	}

	// allocate the new FIP
	if err := AllocateFipRange(newFipRange); err != nil {
		log.Errorf("(UpdateFipRange) Error allocating newFipRange: %s", err.Error())
	}

	return err
}

package kubefip

import (
	"errors"

	KubefipV1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	log "github.com/sirupsen/logrus"
)

func AllocateFipRange(fipRange *KubefipV1.FloatingIPRange) error {
	var err error

	log.Tracef("(AllocateFipRange) fiprangeobj added: [%+v]\n", fipRange)

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

	return err
}

func RemoveFipRange(fipRange *KubefipV1.FloatingIPRange) error {
	var err error

	log.Tracef("(RemoveFipRange) fiprangeobj removed: [%+v]\n", fipRange)

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

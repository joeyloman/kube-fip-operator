package app

import (
	"strings"

	"github.com/joeyloman/kube-fip-operator/pkg/config"

	log "github.com/sirupsen/logrus"
)

func updateLoglevel(kubefipConfig *config.KubefipConfigStruct) {
	currentLogLevel := strings.ToLower(log.GetLevel().String())
	newLogLevel := strings.ToLower(kubefipConfig.LogLevel)

	if strings.EqualFold(currentLogLevel, newLogLevel) {
		return
	}

	level, err := log.ParseLevel(kubefipConfig.LogLevel)
	if err == nil {
		log.Infof("(updateLoglevel) setting loglevel to: %s", kubefipConfig.LogLevel)
		log.SetLevel(level)
	} else {
		log.Infof("(updateLoglevel) level not set in configmap, setting loglevel to: Info")
		log.SetLevel(log.InfoLevel)
	}
}

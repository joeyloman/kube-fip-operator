package metrics

import (
	"fmt"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type appMetricsStruct struct {
	kubefipoperatorFiprangesCapacity  *prometheus.GaugeVec
	kubefipoperatorFiprangesReserved  *prometheus.GaugeVec
	kubefipoperatorGuestclusterStatus *prometheus.GaugeVec
	kubefipoperatorGuestclusterEvents *prometheus.CounterVec
}

type clusterMetricLabels struct {
	guestClusterName     string
	harvesterClustername string
}

var (
	AppMetrics *appMetricsStruct

	LabelFipRangeName = "fiprangename"
	LabelFipRange     = "fiprange"

	LabelFipName = "fipname"
	LabelFip     = "fip"

	LabelGuestClusterName     = "guestclustername"
	LabelHarvesterClusterName = "harvesterclustername"
	LabelHarvesterNetworkName = "harvesternetworkname"
	LabelEvent                = "event"
	LabelStatus               = "status"

	InOperationMode bool = false

	metricsCleanupQueue []clusterMetricLabels
)

const (
	GuestClusterConnectionError      = 1
	KubevipConfigMapManagementError  = 2
	KubevipInstallError              = 3
	KubevipCloudproviderInstallError = 4

	EventApiConnection               = "api_connection"
	EventConfigmapManagement         = "configmap_management"
	EventKubevipInstall              = "kubevip_install"
	EventKubevipCloudproviderInstall = "kubevipcloudprovider_install"

	StatusSuccess = "success"
	StatusError   = "error"

	StatusUp   = 1
	StatusDown = 0
)

func NewMetrics(reg prometheus.Registerer) *appMetricsStruct {
	m := &appMetricsStruct{
		kubefipoperatorFiprangesCapacity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kubefipoperator_fipranges_capacity",
				Help: "Capacity of the Fip ranges",
			},
			[]string{
				LabelFipRangeName,
				LabelFipRange,
				LabelHarvesterClusterName,
				LabelHarvesterNetworkName,
			},
		),
		kubefipoperatorFiprangesReserved: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kubefipoperator_fipranges_reserved",
				Help: "Reserved amount of Fips in a range",
			},
			[]string{
				LabelFipRangeName,
				LabelFipRange,
				LabelHarvesterClusterName,
				LabelHarvesterNetworkName,
			},
		),
		kubefipoperatorGuestclusterStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kubefipoperator_guestcluster_status",
				Help: "Keeps track if the Guest Clusters are 0(down) or 1(up)",
			},
			[]string{
				LabelGuestClusterName,
				LabelHarvesterClusterName,
			},
		),
		kubefipoperatorGuestclusterEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubefipoperator_guestcluster_events",
				Help: "Number of kube-fip-operator guest cluster events",
			},
			[]string{
				LabelGuestClusterName,
				LabelHarvesterClusterName,
				LabelEvent,
				LabelStatus,
			},
		),
	}

	reg.MustRegister(m.kubefipoperatorFiprangesCapacity)
	reg.MustRegister(m.kubefipoperatorFiprangesReserved)
	reg.MustRegister(m.kubefipoperatorGuestclusterStatus)
	reg.MustRegister(m.kubefipoperatorGuestclusterEvents)

	return m
}

func SetFiprangesCapacity(fipRangeName string, fipRange string, harvesterClusterName string, harvesterNetworkName string) {
	_, ipv4Net, err := net.ParseCIDR(fipRange)
	if err != nil {
		log.Errorf("(SetFiprangesCapacity) error while parsing fiprange string to an IPNet object: %s", err)
	}
	fipRangeCapacity := cidr.AddressCount(ipv4Net)

	// the first and the last address in the range are reserved for gateway and broadcast purposes
	fipRangeCapacity = fipRangeCapacity - 2

	log.Debugf("(SetFiprangesCapacity) changing fipranges capacity metric: fipRangeName=%s, fipRange=%s, harvesterClusterName=%s, harvesterNetworkName=%s, fipRangeCapacity=%d",
		fipRangeName, fipRange, harvesterClusterName, harvesterNetworkName, fipRangeCapacity)

	AppMetrics.kubefipoperatorFiprangesCapacity.With(prometheus.Labels{
		LabelFipRangeName:         fipRangeName,
		LabelFipRange:             fipRange,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelHarvesterNetworkName: harvesterNetworkName,
	}).Set(float64(fipRangeCapacity))
}

func IncrementFiprangesReserved(fipRangeName string, fipRange string, harvesterClusterName string, harvesterNetworkName string) {
	log.Debugf("(IncrementFiprangesReserved) incrementing fipranges reserved metric: fipRangeName=%s, fipRange=%s, harvesterClusterName=%s, harvesterNetworkName=%s",
		fipRangeName, fipRange, harvesterClusterName, harvesterNetworkName)

	AppMetrics.kubefipoperatorFiprangesReserved.With(prometheus.Labels{
		LabelFipRangeName:         fipRangeName,
		LabelFipRange:             fipRange,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelHarvesterNetworkName: harvesterNetworkName,
	}).Inc()
}

func DecrementFiprangesReserved(fipRangeName string, fipRange string, harvesterClusterName string, harvesterNetworkName string) {
	log.Debugf("(DecrementFiprangesReserved) decrementing fipranges reserved metric: fipRangeName=%s, fipRange=%s, harvesterClusterName=%s, harvesterNetworkName=%s",
		fipRangeName, fipRange, harvesterClusterName, harvesterNetworkName)

	AppMetrics.kubefipoperatorFiprangesReserved.With(prometheus.Labels{
		LabelFipRangeName:         fipRangeName,
		LabelFipRange:             fipRange,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelHarvesterNetworkName: harvesterNetworkName,
	}).Dec()
}

func RemoveFiprangeMetrics(fipRangeName string, fipRange string, harvesterClusterName string, harvesterNetworkName string) {
	log.Debugf("(RemoveFiprangeMetrics) removing fiprange metrics: fipRangeName=%s, fipRange=%s, harvesterClusterName=%s, harvesterNetworkName=%s",
		fipRangeName, fipRange, harvesterClusterName, harvesterNetworkName)

	AppMetrics.kubefipoperatorFiprangesCapacity.Delete(prometheus.Labels{
		LabelFipRangeName:         fipRangeName,
		LabelFipRange:             fipRange,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelHarvesterNetworkName: harvesterNetworkName,
	})

	AppMetrics.kubefipoperatorFiprangesReserved.Delete(prometheus.Labels{
		LabelFipRangeName:         fipRangeName,
		LabelFipRange:             fipRange,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelHarvesterNetworkName: harvesterNetworkName,
	})
}

func SetGuestClusterStatus(guestClusterName string, harvesterClusterName string, clusterStatus float64) {
	log.Debugf("(SetGuestClusterStatus) changing status metric: guestClusterName=%s, harvesterClusterName=%s, status=%d",
		guestClusterName, harvesterClusterName, int(clusterStatus))

	AppMetrics.kubefipoperatorGuestclusterStatus.With(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
	}).Set(clusterStatus)
}

func IncrementGuestClusterEventsMetric(guestClusterName string, harvesterClusterName string, event string, status string) {
	log.Debugf("(IncrementGuestClusterEventsMetric) incrementing metrics: guestClusterName=%s, harvesterClusterName=%s, event=%s, status=%s",
		guestClusterName, harvesterClusterName, event, status)

	AppMetrics.kubefipoperatorGuestclusterEvents.With(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelEvent:                event,
		LabelStatus:               status,
	}).Inc()
}

func RemoveGuestClusterEventsFromMetrics(guestClusterName string, harvesterClusterName string) {
	log.Debugf("(RemoveGuestClusterEventsFromMetrics) removing metrics: guestClusterName=%s, harvesterClusterName=%s",
		guestClusterName, harvesterClusterName)

	AppMetrics.kubefipoperatorGuestclusterStatus.Delete(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
	})

	AppMetrics.kubefipoperatorGuestclusterEvents.Delete(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelEvent:                EventApiConnection,
		LabelStatus:               StatusError,
	})

	AppMetrics.kubefipoperatorGuestclusterEvents.Delete(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelEvent:                EventApiConnection,
		LabelStatus:               StatusSuccess,
	})

	AppMetrics.kubefipoperatorGuestclusterEvents.Delete(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelEvent:                EventConfigmapManagement,
		LabelStatus:               StatusError,
	})

	AppMetrics.kubefipoperatorGuestclusterEvents.Delete(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelEvent:                EventConfigmapManagement,
		LabelStatus:               StatusSuccess,
	})

	AppMetrics.kubefipoperatorGuestclusterEvents.Delete(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelEvent:                EventKubevipInstall,
		LabelStatus:               StatusError,
	})

	AppMetrics.kubefipoperatorGuestclusterEvents.Delete(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelEvent:                EventKubevipInstall,
		LabelStatus:               StatusSuccess,
	})

	AppMetrics.kubefipoperatorGuestclusterEvents.Delete(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelEvent:                EventKubevipCloudproviderInstall,
		LabelStatus:               StatusError,
	})

	AppMetrics.kubefipoperatorGuestclusterEvents.Delete(prometheus.Labels{
		LabelGuestClusterName:     guestClusterName,
		LabelHarvesterClusterName: harvesterClusterName,
		LabelEvent:                EventKubevipCloudproviderInstall,
		LabelStatus:               StatusSuccess,
	})
}

func AddClusterToMetricsCleanupQueue(guestClusterName string, harvesterClusterName string) {
	log.Debugf("(AddClusterToMetricsCleanupQueue) add guest cluster [%s] and harvester cluster [%s] to the cleanup queue",
		guestClusterName, harvesterClusterName)

	m := clusterMetricLabels{}
	m.guestClusterName = guestClusterName
	m.harvesterClustername = harvesterClusterName

	metricsCleanupQueue = append(metricsCleanupQueue, m)
}

func CleanupMetrics() {
	log.Debugf("(CleanupMetrics) starting cleanup of removed metrics")

	if InOperationMode {
		log.Debugf("(CleanupMetrics) operator mode is running, skipping cleanup session")

		return
	}

	// copy the queue to the in progress queue
	metricsCleanupQueueInProgress := metricsCleanupQueue

	// reset the metricsCleanupQueue queue
	metricsCleanupQueue = nil

	for i := 0; i < len(metricsCleanupQueueInProgress); i++ {
		RemoveGuestClusterEventsFromMetrics(metricsCleanupQueueInProgress[i].guestClusterName, metricsCleanupQueueInProgress[i].harvesterClustername)
	}

	log.Debugf("(CleanupMetrics) finished the cleanup of removed cluster metrics")
}

func InitMetrics(metricsPort int) {
	log.Infof("(InitMetrics) start the init of the metrics")

	// create a non-global registry.
	reg := prometheus.NewRegistry()

	// create new metrics and register them using the custom registry.
	AppMetrics = NewMetrics(reg)

	listenAddress := fmt.Sprintf(":%d", metricsPort)

	// expose metrics and custom registry via an HTTP server
	// using the HandleFor function. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

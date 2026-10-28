package collectors

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/swagner-de/tplinkexporter/clients"
)

type TrafficCollector struct {
	namespace      string
	clients        []clients.TPLINKSwitchClient
	pktMetrics     map[string]*prometheus.GaugeVec
	statusMetrics  map[string]*prometheus.GaugeVec
	settingMetrics map[string]*prometheus.GaugeVec
	upMetric       *prometheus.GaugeVec
	infoMetric     *prometheus.GaugeVec
}

const (
	hostLabel       = "host"
	portnumLabel    = "portnum"
	portstatsSubsys = "portstats"
)

var statusMetricsFields = []string{
	"State",
	"LinkStatus",
}

var pktMetricsFields = []string{
	"TxGoodPkt",
	"TxBadPkt",
	"RxGoodPkt",
	"RxBadPkt",
}

// Values are the switch's raw speed/flow-control codes from PortSettingRpm.
var settingMetricsFields = map[string]string{
	"SpeedConfig":       "Configured port speed/duplex code",
	"SpeedActual":       "Negotiated port speed/duplex code (0 = no link)",
	"FlowControlConfig": "Configured flow control (1 = on)",
	"FlowControlActual": "Active flow control (1 = on)",
}

func NewTrafficCollector(namespace string, switchClients ...clients.TPLINKSwitchClient) *TrafficCollector {
	pktMetrics := make(map[string]*prometheus.GaugeVec)
	for _, name := range pktMetricsFields {
		pktMetrics[name] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: portstatsSubsys,
				Name:      strings.ToLower(name),
				Help:      fmt.Sprintf("Value of the '%s' traffic metric from the switch", name),
			},
			[]string{portnumLabel, hostLabel},
		)
	}
	statusMetrics := make(map[string]*prometheus.GaugeVec)
	for _, name := range statusMetricsFields {
		statusMetrics[name] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: portstatsSubsys,
				Name:      strings.ToLower(name),
				Help:      fmt.Sprintf("Value of the '%s' traffic metric from the switch", name),
			},
			[]string{portnumLabel, hostLabel},
		)
	}
	settingMetrics := make(map[string]*prometheus.GaugeVec)
	for name, help := range settingMetricsFields {
		settingMetrics[name] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: portstatsSubsys,
				Name:      strings.ToLower(name),
				Help:      help,
			},
			[]string{portnumLabel, hostLabel},
		)
	}
	upMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Whether the last scrape of the switch succeeded (1) or failed (0)",
		},
		[]string{hostLabel},
	)
	infoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "switch_info",
			Help:      "Switch device information; constant 1 with metadata labels",
		},
		[]string{hostLabel, "name", "mac", "ip", "firmware", "hardware"},
	)
	return &TrafficCollector{
		namespace:      namespace,
		clients:        switchClients,
		pktMetrics:     pktMetrics,
		statusMetrics:  statusMetrics,
		settingMetrics: settingMetrics,
		upMetric:       upMetric,
		infoMetric:     infoMetric,
	}
}

func (c *TrafficCollector) Collect(ch chan<- prometheus.Metric) {
	// Reset so stale values from ports/switches that stopped reporting don't linger.
	for name := range c.pktMetrics {
		c.pktMetrics[name].Reset()
	}
	for name := range c.statusMetrics {
		c.statusMetrics[name].Reset()
	}
	for name := range c.settingMetrics {
		c.settingMetrics[name].Reset()
	}
	c.upMetric.Reset()
	c.infoMetric.Reset()

	for _, client := range c.clients {
		host := client.GetHost()
		stats, err := client.GetPortStats()
		if err != nil {
			log.Printf("Error while collecting traffic statistics for %s: %v\n", host, err)
			c.upMetric.With(prometheus.Labels{hostLabel: host}).Set(0)
			continue
		}
		c.upMetric.With(prometheus.Labels{hostLabel: host}).Set(1)
		for portnum := range stats {
			labels := prometheus.Labels{portnumLabel: strconv.Itoa(portnum + 1), hostLabel: host}
			for name, value := range stats[portnum].PktCount {
				c.pktMetrics[name].With(labels).Set(float64(value))
			}
			c.statusMetrics["State"].With(labels).Set(float64(stats[portnum].State))
			c.statusMetrics["LinkStatus"].With(labels).Set(float64(stats[portnum].LinkStatus))
		}

		// Port settings and system info are best-effort: a failure here (often a
		// transient session drop) must not fail the whole scrape or flip 'up'.
		c.collectPortSettings(host, client)
		c.collectSystemInfo(host, client)
	}

	for name := range c.pktMetrics {
		c.pktMetrics[name].Collect(ch)
	}
	for name := range c.statusMetrics {
		c.statusMetrics[name].Collect(ch)
	}
	for name := range c.settingMetrics {
		c.settingMetrics[name].Collect(ch)
	}
	c.upMetric.Collect(ch)
	c.infoMetric.Collect(ch)
}

func (c *TrafficCollector) collectPortSettings(host string, client clients.TPLINKSwitchClient) {
	settings, err := client.GetPortSettings()
	if err != nil {
		log.Printf("Error while collecting port settings for %s: %v\n", host, err)
		return
	}
	for portnum := range settings {
		labels := prometheus.Labels{portnumLabel: strconv.Itoa(portnum + 1), hostLabel: host}
		c.settingMetrics["SpeedConfig"].With(labels).Set(float64(settings[portnum].SpeedConfig))
		c.settingMetrics["SpeedActual"].With(labels).Set(float64(settings[portnum].SpeedActual))
		c.settingMetrics["FlowControlConfig"].With(labels).Set(float64(settings[portnum].FlowControlConfig))
		c.settingMetrics["FlowControlActual"].With(labels).Set(float64(settings[portnum].FlowControlActual))
	}
}

func (c *TrafficCollector) collectSystemInfo(host string, client clients.TPLINKSwitchClient) {
	info, err := client.GetSystemInfo()
	if err != nil {
		log.Printf("Error while collecting system info for %s: %v\n", host, err)
		return
	}
	c.infoMetric.With(prometheus.Labels{
		hostLabel:  host,
		"name":     info.Name,
		"mac":      info.MAC,
		"ip":       info.IP,
		"firmware": info.Firmware,
		"hardware": info.Hardware,
	}).Set(1)
}

func (c *TrafficCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, name := range pktMetricsFields {
		c.pktMetrics[name].Describe(ch)
	}
	for _, name := range statusMetricsFields {
		c.statusMetrics[name].Describe(ch)
	}
	c.upMetric.Describe(ch)
}

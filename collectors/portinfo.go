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
	namespace     string
	clients       []clients.TPLINKSwitchClient
	pktMetrics    map[string]*prometheus.GaugeVec
	statusMetrics map[string]*prometheus.GaugeVec
	upMetric      *prometheus.GaugeVec
}

const (
	hostLabel    = "host"
	portnumLabel = "portnum"
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

func NewTrafficCollector(namespace string, switchClients ...clients.TPLINKSwitchClient) *TrafficCollector {
	pktMetrics := make(map[string]*prometheus.GaugeVec)
	for _, name := range pktMetricsFields {
		pktMetrics[name] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "portstats",
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
				Subsystem: "portstats",
				Name:      strings.ToLower(name),
				Help:      fmt.Sprintf("Value of the '%s' traffic metric from the switch", name),
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
	return &TrafficCollector{
		namespace:     namespace,
		clients:       switchClients,
		pktMetrics:    pktMetrics,
		statusMetrics: statusMetrics,
		upMetric:      upMetric,
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
	c.upMetric.Reset()

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
	}

	for name := range c.pktMetrics {
		c.pktMetrics[name].Collect(ch)
	}
	for name := range c.statusMetrics {
		c.statusMetrics[name].Collect(ch)
	}
	c.upMetric.Collect(ch)
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

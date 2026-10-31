package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/swagner-de/tplinkexporter/clients"
	"github.com/swagner-de/tplinkexporter/collectors"
	"github.com/swagner-de/tplinkexporter/config"
)

func main() {
	var (
		configFile = kingpin.Flag("config",
			"Path to a YAML config file. When set, credentials are read from the "+
				"file and --host/--password are not required.").String()
		host     = kingpin.Flag("host", "Host of target tplink easysmart switch.").String()
		username = kingpin.Flag("username", "Username for switch GUI login").Default("admin").String()
		password = kingpin.Flag("password", "Password for switch GUI login").String()
		port     = kingpin.Flag("port", "Metrics port to listen on for prometheus scrapes").Default("9717").Int()
		timeout  = kingpin.Flag("timeout", "HTTP timeout (seconds) for switch requests").Default("10").Int()
	)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	cfg, err := resolveConfig(*configFile, *host, *username, *password, *port, *timeout)
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	timeoutDuration := time.Duration(cfg.TimeoutSeconds) * time.Second
	switchClients := make([]clients.TPLINKSwitchClient, 0, len(cfg.Switches))
	for _, sw := range cfg.Switches {
		switchClients = append(switchClients, clients.NewTPLinkSwitch(sw.Host, sw.Username, sw.Password, timeoutDuration))
	}

	trafficCollector := collectors.NewTrafficCollector("tplinkexporter", switchClients...)
	prometheus.MustRegister(trafficCollector)
	http.Handle("/metrics", promhttp.Handler())

	// /healthz is a cheap liveness/readiness endpoint that does not scrape the
	// switches. Probing /metrics triggers a live switch scrape, so a slow or
	// unreachable switch would fail the probe and get the pod killed.
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":" + strconv.Itoa(cfg.ListenPort)
	log.Printf("Beginning to serve on port %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// resolveConfig builds the config from either a config file or the legacy flags.
func resolveConfig(configFile, host, username, password string, port, timeout int) (*config.Config, error) {
	if configFile != "" {
		return config.Load(configFile)
	}

	if host == "" {
		return nil, errors.New("either --config or --host must be provided")
	}
	if password == "" {
		return nil, errors.New("--password is required when --config is not set")
	}
	return &config.Config{
		ListenPort:     port,
		TimeoutSeconds: timeout,
		Switches: []config.Switch{
			{Host: host, Username: username, Password: password},
		},
	}, nil
}

# TPLink EasySmart Switch Exporter

Prometheus exporter for TP-Link **Easy Smart** switches (TL-SG108E / TL-SG108PE
family). These devices have no SNMP, so the exporter scrapes the switch web UI
and exposes per-port packet counters and link status on `/metrics`.

This is the `github.com/swagner-de/tplinkexporter` fork, hardened for GitOps
deployment: it adds config-file support (so credentials can be supplied from a
mounted secret), a cookie-jar HTTP client with timeouts, and a scrape-health
metric.

Tested on:

- TL-SG108E v3/v4
- TL-SG108PE (8-port, 4 PoE)

Should work on other switches in the same family (e.g. TL-SG116E), but untested.

## Usage

### With a config file (recommended)

Create a `config.yaml` (see [`config.example.yaml`](config.example.yaml)):

```yaml
listen_port: 9717      # optional, default 9717
timeout_seconds: 10    # optional, HTTP timeout, default 10
switches:
  - host: 192.0.2.10
    username: admin      # optional, default "admin"
    password: "..."      # supply via a mounted secret in production
```

Multiple switches may be listed; each is exported with its own `host` label.

Run it:

```bash
tplinkexporter --config config.yaml
```

The config is validated on startup and the exporter fails fast with a clear
error if the file is missing/malformed or a required field is empty.
Credentials are never logged.

### With CLI flags (legacy)

Backward-compatible flag mode is still supported for a single switch:

```bash
go run . --host 10.0.0.3 --username admin --password admin
```

The default username and password for these switches is `admin` / `admin`.
When `--config` is set, `--host`/`--password` are not required and config-file
values are used.

### With Docker

```bash
docker run -it -p 9717:9717 \
  -v "$PWD/config.yaml:/config/config.yaml:ro" \
  ghcr.io/swagner-de/tplinkexporter --config /config/config.yaml
```

The image is published multi-arch (`linux/amd64,linux/arm64`) at
`ghcr.io/swagner-de/tplinkexporter`, runs as non-root uid `65532`, and works on
a read-only root filesystem.

### On Kubernetes with a secret file

Mount a secret containing the rendered config at `/config/config.yaml` and run
the exporter with `--config /config/config.yaml`. This lets you source the
switch password from a `Secret`/`ExternalSecret` without ever putting it on the
command line (and therefore in Git). Run the container with
`runAsNonRoot: true`, `runAsUser: 65532`, `readOnlyRootFilesystem: true`, all
capabilities dropped, and `seccompProfile: RuntimeDefault`.

## Security note (CVE-2017-17746)

These switches authorise any request coming from an IP that has an active login
session, rather than binding authorisation to the session cookie. The exporter
uses a cookie jar and sends the session cookie explicitly, but you should still
ensure the **exporter pod is the only client talking to the switch from its
source IP**. Do not share that source IP with other clients.

## Grafana Dashboard

Basic Grafana dashboard using the exporter: https://grafana.com/grafana/dashboards/12517

## Metrics exported on `:9717/metrics`

```
tplinkexporter_up{host="host"}                               # 1 if the last scrape of the switch succeeded, else 0
tplinkexporter_portstats_state{portnum="1"-"8",host="host"}
tplinkexporter_portstats_linkstatus{portnum="1"-"8",host="host"}
tplinkexporter_portstats_rxgoodpkt{portnum="1"-"8",host="host"}
tplinkexporter_portstats_rxbadpkt{portnum="1"-"8",host="host"}
tplinkexporter_portstats_txgoodpkt{portnum="1"-"8",host="host"}
tplinkexporter_portstats_txbadpkt{portnum="1"-"8",host="host"}
```

Packet counters are only reported for enabled ports (switch `state == 1`).

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md). Common commands:

```bash
make test   # unit tests with coverage
make lint   # golangci-lint
make build  # build bin/exporter
```

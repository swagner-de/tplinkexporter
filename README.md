# TPLink EasySmart Switch Exporter

Prometheus exporter for TP-Link **Easy Smart** switches (TL-SG108E / TL-SG108PE
family). These devices have no SNMP, so the exporter scrapes the switch web UI
and exposes per-port packet counters and link status on `/metrics`.

## About this fork

This project started as a fork of
[`thelastguardian/tplinkexporter`](https://github.com/thelastguardian/tplinkexporter),
which appears unmaintained: the last functional change landed in 2022, the only
later commit was a one-line Dockerfile tweak, and open issues and pull requests
have sat unanswered. Rather than wait on upstream, this fork carries the project
forward and has since diverged — it is published under a renamed Go module
(`github.com/swagner-de/tplinkexporter`) and its own release line, and is no
longer linked to the original repository.

Changes over upstream:

- Config-file support, so credentials can be mounted from a Secret instead of
  passed on the command line.
- Correct session handling: a single HTTP client with a cookie jar, plus request
  timeouts (upstream relied on a switch firmware quirk, see the CVE note below).
- A `tplinkexporter_up` scrape-health metric for alerting.
- Support for scraping multiple switches from one exporter.
- Modern toolchain, linting, tests, a distroless non-root image, and automated
  multi-arch releases.

Tested against:

- TL-SG108E v3/v4
- TL-SG108PE (8-port, 4 PoE)

Other switches in the same Easy Smart family (e.g. TL-SG116E) are expected to
work but are untested here.

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

### On Kubernetes with Helm

A Helm chart lives under [`chart/`](chart/). It renders the non-secret config
(host, username, port) into a ConfigMap and reads each switch's password from a
file mounted out of a Secret — so the password never lands in the ConfigMap.

For each switch you supply **exactly one** password source:

- `password` — an inline value. The chart creates a Secret for that switch.
  Supply it via `--set` or your CI secret store; **never commit a real password**
  to a values file.
- `existingSecret` — the name of a Secret you manage (e.g. produced by an
  `ExternalSecret`). `existingSecretKey` selects the key within it (default
  `password`).

Switches may mix the two sources. Install from a local checkout (automated OCI
publishing is not wired up yet):

```bash
# Switch with a chart-managed Secret (password supplied at install time):
helm install tplink ./chart \
  --set config.switches[0].host=switch1.example.lan \
  --set config.switches[0].password="$SWITCH1_PASSWORD"

# Switch referencing an existing Secret:
helm install tplink ./chart \
  --set config.switches[0].host=switch1.example.lan \
  --set config.switches[0].existingSecret=switch1-credentials
```

Set `serviceMonitor.enabled=true` if you run the Prometheus Operator. See
[`chart/values.yaml`](chart/values.yaml) for all options. The pod runs non-root
(uid `65532`) on a read-only root filesystem by default.

The chart's `appVersion` tracks the exporter image tag and is kept current by
Renovate.

### On Kubernetes with a raw config secret

Alternatively, mount a Secret containing the full rendered `config.yaml` at
`/config/config.yaml` and run the exporter with `--config /config/config.yaml`.
Run the container with `runAsNonRoot: true`, `runAsUser: 65532`,
`readOnlyRootFilesystem: true`, all capabilities dropped, and
`seccompProfile: RuntimeDefault`.

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
tplinkexporter_up{host}                                    # 1 if the last scrape succeeded, else 0
tplinkexporter_switch_info{host,name,mac,ip,firmware,hardware}  # constant 1; device metadata in labels

tplinkexporter_portstats_state{portnum,host}               # 1 = port enabled
tplinkexporter_portstats_linkstatus{portnum,host}          # link speed/state code reported by the switch
tplinkexporter_portstats_rxgoodpkt{portnum,host}
tplinkexporter_portstats_rxbadpkt{portnum,host}
tplinkexporter_portstats_txgoodpkt{portnum,host}
tplinkexporter_portstats_txbadpkt{portnum,host}

tplinkexporter_portstats_speedconfig{portnum,host}         # configured speed/duplex code
tplinkexporter_portstats_speedactual{portnum,host}         # negotiated speed/duplex code (0 = no link)
tplinkexporter_portstats_flowcontrolconfig{portnum,host}   # 1 = flow control configured on
tplinkexporter_portstats_flowcontrolactual{portnum,host}   # 1 = flow control active
```

`portnum` is 1-based and covers every port the switch reports (an 8-port model
reports 10 entries, the last two being internal). Packet counters are only
emitted for enabled ports (`state == 1`); disabled ports expose `state` and
`linkstatus` only. The `speed*`/`flowcontrol*` and `switch_info` series are
best-effort — a transient failure to read those pages is logged but does not
fail the scrape or flip `up`.

The speed code follows the switch's web UI: `0` = no link, `2` = 10M half,
`3` = 10M full, `4` = 100M half, `5` = 100M full, `6` = 1000M full.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md). Common commands:

```bash
make test   # unit tests with coverage
make lint   # golangci-lint
make build  # build bin/exporter
```

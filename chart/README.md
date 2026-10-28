# tplinkexporter

Chart version: 0.2.0 <!-- x-release-please-version -->
![AppVersion: 0.4.0](https://img.shields.io/badge/AppVersion-0.4.0-informational?style=flat-square)  ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)
Prometheus exporter for TP-Link Easy Smart switches
**Homepage:** <https://github.com/swagner-de/tplinkexporter>

## Features
- Scrapes per-port packet counters and link status from TP-Link Easy Smart switches
- Scrapes multiple switches from one exporter, each labelled by host
- Exposes a `tplinkexporter_up` scrape-health metric for alerting

## Install

<!-- x-release-please-start-version -->
```bash
# Switch with a chart-managed Secret (password supplied at install time):
helm install tplink oci://ghcr.io/swagner-de/charts/tplinkexporter \
  --version 0.2.0 \
  --set config.switches[0].host=switch1.example.lan \
  --set config.switches[0].password="$SWITCH1_PASSWORD"

# Switch referencing an existing Secret:
helm install tplink oci://ghcr.io/swagner-de/charts/tplinkexporter \
  --version 0.2.0 \
  --set config.switches[0].host=switch1.example.lan \
  --set config.switches[0].existingSecret=switch1-credentials
```
<!-- x-release-please-end -->

Each switch needs exactly one password source: an inline `password` (the chart
creates a Secret for it) or an `existingSecret` reference. Never commit a real
inline password to a values file.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules for pod scheduling. |
| config.listenPort | int | `9717` | Metrics port the exporter listens on. |
| config.switches | list | `[{"existingSecret":"","existingSecretKey":"password","host":"switch1.example.lan","password":"changeme","username":"admin"}]` | Switches to scrape. Each entry needs a host and exactly one password source: an inline `password` (the chart creates a Secret) or an `existingSecret` reference (with `existingSecretKey`, default `password`). Supply inline passwords via --set or CI, never commit real values. |
| config.timeoutSeconds | int | `10` | HTTP timeout in seconds for each switch request. |
| fullnameOverride | string | `""` | Override the fully qualified release name. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. |
| image.repository | string | `"ghcr.io/swagner-de/tplinkexporter"` | Container image repository. |
| image.tag | string | `""` | Image tag. Defaults to the chart's appVersion when empty. |
| imagePullSecrets | list | `[]` | Image pull secrets for private registries. |
| nameOverride | string | `""` | Override the chart name used in resource names. |
| nodeSelector | object | `{}` | Node selector for pod scheduling. |
| podAnnotations | object | `{}` | Extra annotations for the pod. |
| podSecurityContext | object | `{"runAsGroup":65532,"runAsNonRoot":true,"runAsUser":65532,"seccompProfile":{"type":"RuntimeDefault"}}` | Pod-level security context. |
| replicaCount | int | `1` | Number of exporter replicas. One is sufficient; the exporter is stateless. |
| resources | object | `{"limits":{"memory":"64Mi"},"requests":{"cpu":"10m","memory":"32Mi"}}` | Container resource requests and limits. |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true}` | Container-level security context. |
| service.port | int | `9717` | Service port exposing /metrics. |
| service.type | string | `"ClusterIP"` | Kubernetes Service type. |
| serviceMonitor.enabled | bool | `false` | Create a Prometheus Operator ServiceMonitor. |
| serviceMonitor.interval | string | `"30s"` | Scrape interval. |
| serviceMonitor.labels | object | `{}` | Extra labels for the ServiceMonitor (e.g. to match a Prometheus selector). |
| serviceMonitor.scrapeTimeout | string | `"10s"` | Per-scrape timeout. |
| tolerations | list | `[]` | Tolerations for pod scheduling. |

## Security
- `runAsNonRoot: true`, UID/GID 65532
- `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true`
- All capabilities dropped
- Seccomp profile: `RuntimeDefault`

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| swagner-de | <swagner-de@users.noreply.github.com> |  |

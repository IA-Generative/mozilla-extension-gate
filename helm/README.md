# mozilla-extension-gate

![Version: 0.1.2](https://img.shields.io/badge/Version-0.1.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.2.0](https://img.shields.io/badge/AppVersion-1.2.0-informational?style=flat-square)

A Helm chart for mozilla-extension-gate application

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| ArnaudTA | <arnaud@example.com> |  |

## Values

### Servicename

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| serviceMonitor.annotations | object | `{}` | Prometheus ServiceMonitor annotations. |
| serviceMonitor.enabled | bool | `true` | Enable a prometheus ServiceMonitor. |
| serviceMonitor.endpoints[0] | object | `{"honorLabels":false,"interval":"30s","metricRelabelings":[],"path":"/metrics","relabelings":[],"scheme":"","scrapeTimeout":"10s","selector":{},"tlsConfig":{}}` | Prometheus ServiceMonitor interval. |
| serviceMonitor.endpoints[0].honorLabels | bool | `false` | When true, honorLabels preserves the metric’s labels when they collide with the target’s labels. |
| serviceMonitor.endpoints[0].metricRelabelings | list | `[]` | Prometheus MetricRelabelConfigs to apply to samples before ingestion. |
| serviceMonitor.endpoints[0].path | string | `"/metrics"` | Path used by the Prometheus ServiceMonitor to scrape metrics. |
| serviceMonitor.endpoints[0].relabelings | list | `[]` | Prometheus RelabelConfigs to apply to samples before scraping. |
| serviceMonitor.endpoints[0].scheme | string | `""` | Prometheus ServiceMonitor scheme. |
| serviceMonitor.endpoints[0].scrapeTimeout | string | `"10s"` | Prometheus ServiceMonitor scrapeTimeout. If empty, Prometheus uses the global scrape timeout unless it is less than the target's scrape interval value in which the latter is used. |
| serviceMonitor.endpoints[0].selector | object | `{}` | Prometheus ServiceMonitor selector. |
| serviceMonitor.endpoints[0].tlsConfig | object | `{}` | Prometheus ServiceMonitor tlsConfig. |
| serviceMonitor.labels | object | `{}` | Prometheus ServiceMonitor labels. |

### Other Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| config.artifactsPrefix | string | `"artifacts"` |  |
| config.github.repository | string | `""` |  |
| config.github.syncInterval | string | `"5m"` |  |
| config.manifestsPrefix | string | `"manifests"` |  |
| config.s3.accessKeyId | string | `""` |  |
| config.s3.bucket | string | `""` |  |
| config.s3.endpoint | string | `""` |  |
| config.s3.region | string | `"us-east-1"` |  |
| config.s3.secretAccessKey | string | `""` |  |
| config.server.artifactBaseUrl | string | `""` |  |
| config.server.basePath | string | `"/"` |  |
| config.server.env | string | `"production"` |  |
| config.server.instanceId | string | `"gateway-instance"` |  |
| config.server.port | string | `"8080"` |  |
| config.server.proxyArtifactEnabled | bool | `false` |  |
| config.server.publicUrl | string | `"http://localhost:8080"` |  |
| config.server.version | string | `"1.0"` |  |
| env | list | `[]` |  |
| extraCAs | object | `{}` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/ia-generative/mozilla-extension-gate/server"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"Prefix"` |  |
| ingress.tls | list | `[]` |  |
| initContainer.resources.limits.cpu | string | `"500m"` |  |
| initContainer.resources.limits.memory | string | `"512Mi"` |  |
| initContainer.resources.requests.cpu | string | `"250m"` |  |
| initContainer.resources.requests.memory | string | `"256Mi"` |  |
| nameOverride | string | `""` |  |
| nodeAffinity | object | `{}` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| proxy.httpProxy | string | `""` |  |
| proxy.httpsProxy | string | `""` |  |
| proxy.noProxy | string | `""` |  |
| replicaCount | int | `1` |  |
| resources.limits.cpu | string | `"500m"` |  |
| resources.limits.memory | string | `"512Mi"` |  |
| resources.requests.cpu | string | `"250m"` |  |
| resources.requests.memory | string | `"256Mi"` |  |
| securityContext | object | `{}` |  |
| service.annotations | object | `{}` |  |
| service.port | int | `80` |  |
| service.targetPort | int | `8080` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tolerations | list | `[]` |  |
| volumeMounts | list | `[]` |  |
| volumes | list | `[]` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)

"""
Common labels and configuration for the collector-agent resources
"""

import schema

# Get command-line options
_image = option("config.image")
_namespace = option("config.namespace")
_clusterName = option("config.clusterName")
_serverURL = option("config.serverURL")
_clusterEndpoint = option("config.clusterEndpoint")
_reportInterval = option("config.reportInterval")
_heartbeatInterval = option("config.heartbeatInterval")
_namespaces = option("config.namespaces")
_collectorToken = option("config.collectorToken")

config: schema.CollectorAgent = schema.CollectorAgent {
    if _image:
        image = _image
    if _namespace:
        namespace = _namespace
    if _clusterName:
        clusterName = _clusterName
    if _serverURL:
        serverURL = _serverURL
    if _clusterEndpoint:
        clusterEndpoint = _clusterEndpoint
    if _reportInterval:
        reportInterval = _reportInterval
    if _heartbeatInterval:
        heartbeatInterval = _heartbeatInterval
    if _namespaces:
        namespaces = _namespaces
    if _collectorToken:
        collectorToken = _collectorToken
}

commonLabels = {
    "app.kubernetes.io/name": config.name
    "app.kubernetes.io/instance": config.name
    "app.kubernetes.io/component": "cluster-collector"
    "app.kubernetes.io/part-of": "run-things"
    **config.labels
}

selectorLabels = {
    "app.kubernetes.io/name": config.name
    "app.kubernetes.io/instance": config.name
}

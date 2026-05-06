"""
Schema definition for the run-things collector-agent deployment configuration
"""

schema CollectorAgent:
    """Configuration for the run-things collector-agent (per-cluster informer)"""

    # Metadata
    name: str = "run-things-collector"
    namespace: str = "run-things-collector"
    clusterName: str = "default"

    # Image
    image: str = "ghcr.io/stuttgart-things/run-things-collector-agent:v0.1.0"
    imagePullPolicy: str = "Always"

    # Replicas and resources
    replicas: int = 1
    cpuRequest: str = "20m"
    cpuLimit: str = "100m"
    memoryRequest: str = "32Mi"
    memoryLimit: str = "128Mi"

    # Connection to run-things server (collector ingest port)
    serverURL: str = "http://run-things.run-things.svc.cluster.local:50051"
    clusterEndpoint: str = ""

    # Reporting cadence (Go duration strings, e.g. "60s", "5m")
    reportInterval: str = "60s"
    heartbeatInterval: str = "30s"
    httpTimeout: str = "15s"

    # Optional comma-separated namespace allow list. Empty = cluster-wide.
    namespaces: str = ""

    # Bearer token shared with the run-things server. Empty disables auth.
    # In production wire this up via a Secret + envFrom — never inline literals here.
    collectorToken: str = ""

    # Labels and annotations
    labels: {str:str} = {}
    annotations: {str:str} = {}

    check:
        replicas >= 0, "replicas must be >= 0"
        name != "", "name must not be empty"
        namespace != "", "namespace must not be empty"
        clusterName != "", "clusterName must not be empty"
        serverURL != "", "serverURL must not be empty"

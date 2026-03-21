"""
Schema definition for weRunThings deployment configuration
"""

schema WeRunThings:
    """Configuration for weRunThings Kubernetes deployment"""

    # Metadata
    name: str = "werunthings"
    namespace: str = "werunthings"

    # Image
    image: str = "ghcr.io/stuttgart-things/weRunThings:v0.1.0"
    imagePullPolicy: str = "Always"

    # Replicas and resources
    replicas: int = 1
    cpuRequest: str = "50m"
    cpuLimit: str = "200m"
    memoryRequest: str = "64Mi"
    memoryLimit: str = "256Mi"

    # Ports
    grpcPort: int = 50051
    httpPort: int = 8080
    grpcServicePort: int = 80
    httpServicePort: int = 8080

    # Config
    loadConfigFrom: str = "cr"
    configLocation: str = "werunthings"
    configName: str = "portal-labul"
    serverPort: str = "50051"

    # Service
    serviceType: str = "ClusterIP"

    # HTTPRoute (Gateway API)
    httpRouteEnabled: bool = False
    httpRouteParentRefName: str = ""
    httpRouteParentRefNamespace: str = ""
    httpRouteHostname: str = ""
    httpRouteAnnotations: {str:str} = {}

    # RBAC
    servicePortalApiGroup: str = "github.stuttgart-things.com"

    # Labels and annotations
    labels: {str:str} = {}
    annotations: {str:str} = {}

    check:
        replicas >= 0, "replicas must be >= 0"
        grpcPort >= 1 and grpcPort <= 65535, "grpcPort must be valid"
        httpPort >= 1 and httpPort <= 65535, "httpPort must be valid"
        name != "", "name must not be empty"
        namespace != "", "namespace must not be empty"

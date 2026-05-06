"""
Deployment for the run-things collector-agent
"""

import labels

config = labels.config

deployment = {
    apiVersion: "apps/v1"
    kind: "Deployment"
    metadata: {
        name: config.name
        namespace: config.namespace
        labels: labels.commonLabels
    }
    spec: {
        replicas: config.replicas
        selector: {
            matchLabels: labels.selectorLabels
        }
        strategy: {
            type: "RollingUpdate"
            rollingUpdate: {
                maxSurge: 1
                maxUnavailable: 0
            }
        }
        template: {
            metadata: {
                labels: labels.commonLabels
            }
            spec: {
                serviceAccountName: config.name
                containers: [
                    {
                        name: config.name
                        image: config.image
                        imagePullPolicy: config.imagePullPolicy
                        env: [
                            {name: "CLUSTER_NAME", value: config.clusterName}
                            {name: "SERVER_URL", value: config.serverURL}
                            {name: "CLUSTER_ENDPOINT", value: config.clusterEndpoint}
                            {name: "REPORT_INTERVAL", value: config.reportInterval}
                            {name: "HEARTBEAT_INTERVAL", value: config.heartbeatInterval}
                            {name: "HTTP_TIMEOUT", value: config.httpTimeout}
                            {name: "NAMESPACES", value: config.namespaces}
                            {name: "COLLECTOR_TOKEN", value: config.collectorToken}
                        ]
                        resources: {
                            requests: {
                                cpu: config.cpuRequest
                                memory: config.memoryRequest
                            }
                            limits: {
                                cpu: config.cpuLimit
                                memory: config.memoryLimit
                            }
                        }
                        securityContext: {
                            readOnlyRootFilesystem: True
                            runAsNonRoot: True
                            allowPrivilegeEscalation: False
                            capabilities: {
                                drop: ["ALL"]
                            }
                        }
                    }
                ]
                securityContext: {
                    runAsNonRoot: True
                    runAsUser: 65532
                    runAsGroup: 65532
                    fsGroup: 65532
                    seccompProfile: {
                        type: "RuntimeDefault"
                    }
                }
            }
        }
    }
}

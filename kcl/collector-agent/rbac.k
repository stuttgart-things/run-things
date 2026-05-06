"""
ClusterRole + ClusterRoleBinding so the collector-agent can list workloads
across the whole cluster (read-only).
"""

import labels

config = labels.config

clusterRole = {
    apiVersion: "rbac.authorization.k8s.io/v1"
    kind: "ClusterRole"
    metadata: {
        name: config.name
        labels: labels.commonLabels
    }
    rules: [
        {
            apiGroups: ["apps"]
            resources: ["deployments", "statefulsets", "daemonsets"]
            verbs: ["get", "list", "watch"]
        }
        {
            apiGroups: [""]
            resources: ["services", "namespaces"]
            verbs: ["get", "list", "watch"]
        }
        {
            apiGroups: ["networking.k8s.io"]
            resources: ["ingresses"]
            verbs: ["get", "list", "watch"]
        }
    ]
}

clusterRoleBinding = {
    apiVersion: "rbac.authorization.k8s.io/v1"
    kind: "ClusterRoleBinding"
    metadata: {
        name: config.name
        labels: labels.commonLabels
    }
    subjects: [
        {
            kind: "ServiceAccount"
            name: config.name
            namespace: config.namespace
        }
    ]
    roleRef: {
        kind: "ClusterRole"
        name: config.name
        apiGroup: "rbac.authorization.k8s.io"
    }
}

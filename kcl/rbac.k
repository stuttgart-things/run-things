"""
RBAC resources for weRunThings (Role + RoleBinding for ServicePortal CRs)
"""

import labels

config = labels.config

role = {
    apiVersion: "rbac.authorization.k8s.io/v1"
    kind: "Role"
    metadata: {
        name: config.name
        namespace: config.namespace
        labels: labels.commonLabels
    }
    rules: [
        {
            apiGroups: [config.servicePortalApiGroup]
            resources: ["serviceportals"]
            verbs: ["get", "list", "watch", "create", "update"]
        }
    ]
}

roleBinding = {
    apiVersion: "rbac.authorization.k8s.io/v1"
    kind: "RoleBinding"
    metadata: {
        name: config.name
        namespace: config.namespace
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
        kind: "Role"
        name: config.name
        apiGroup: "rbac.authorization.k8s.io"
    }
}

"""
Main entry point for the run-things collector-agent Kubernetes manifests.
Imports all resources and exports manifests.
"""

import namespace
import serviceaccount
import rbac
import deploy

manifests = [
    m for m in [
        namespace.namespace
        serviceaccount.serviceAccount
        rbac.clusterRole
        rbac.clusterRoleBinding
        deploy.deployment
    ] if m
]

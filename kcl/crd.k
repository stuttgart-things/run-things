"""
CustomResourceDefinition for ServicePortal
"""

servicePortalCRD = {
    apiVersion: "apiextensions.k8s.io/v1"
    kind: "CustomResourceDefinition"
    metadata: {
        name: "serviceportals.github.stuttgart-things.com"
    }
    spec: {
        group: "github.stuttgart-things.com"
        names: {
            kind: "ServicePortal"
            plural: "serviceportals"
            singular: "serviceportal"
            shortNames: ["sp"]
        }
        scope: "Namespaced"
        versions: [
            {
                name: "v1"
                served: True
                storage: True
                schema: {
                    openAPIV3Schema: {
                        type: "object"
                        properties: {
                            spec: {
                                type: "object"
                                properties: {
                                    services: {
                                        type: "array"
                                        items: {
                                            type: "object"
                                            properties: {
                                                name: { type: "string" }
                                                description: { type: "string" }
                                                category: { type: "string" }
                                                url: { type: "string" }
                                                logoURL: { type: "string" }
                                                icon: { type: "string" }
                                                tags: {
                                                    type: "array"
                                                    items: { type: "string" }
                                                }
                                                healthCheck: {
                                                    type: "object"
                                                    properties: {
                                                        enabled: { type: "boolean" }
                                                        interval: { type: "integer" }
                                                        method: { type: "string" }
                                                        expectedStatus: { type: "integer" }
                                                        expectedBody: { type: "string" }
                                                        tlsCheck: { type: "boolean" }
                                                        timeout: { type: "integer" }
                                                    }
                                                }
                                            }
                                            required: ["name", "url"]
                                        }
                                        description: "List of services to monitor and display in the portal"
                                    }
                                }
                                required: ["services"]
                            }
                        }
                    }
                }
            }
        ]
    }
}

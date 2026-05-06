/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// newDynamicClient builds a kubernetes dynamic client. It tries in-cluster
// config first (the common case when the agent runs as a Pod) and falls back
// to the kubeconfig path for local development.
func newDynamicClient(kubeconfigPath string) (dynamic.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}
	return dynamic.NewForConfig(cfg)
}

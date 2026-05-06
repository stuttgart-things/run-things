/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	gvrDeployment  = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	gvrStatefulSet = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
	gvrDaemonSet   = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}
	gvrService     = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	gvrIngress     = schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
)

// Discoverer collects workload inventory from a Kubernetes cluster.
type Discoverer struct {
	Client      dynamic.Interface
	Namespaces  []string // empty = all namespaces
	ClusterName string
}

// Snapshot returns the current cluster inventory.
func (d *Discoverer) Snapshot(ctx context.Context) (ClusterInventory, error) {
	inv := ClusterInventory{
		ClusterName: d.ClusterName,
		LastUpdated: time.Now().UTC(),
	}

	collect := func(gvr schema.GroupVersionResource, kind string, withReplicas bool) []WorkloadInfo {
		items, err := d.list(ctx, gvr)
		if err != nil {
			log.Printf("list %s failed: %v", kind, err)
			return nil
		}
		out := make([]WorkloadInfo, 0, len(items))
		for _, it := range items {
			out = append(out, toWorkload(it, kind, withReplicas))
		}
		return out
	}

	inv.Deployments = collect(gvrDeployment, "Deployment", true)
	inv.StatefulSets = collect(gvrStatefulSet, "StatefulSet", true)
	inv.DaemonSets = collect(gvrDaemonSet, "DaemonSet", true)
	inv.Services = collect(gvrService, "Service", false)
	inv.Ingresses = collect(gvrIngress, "Ingress", false)
	return inv, nil
}

func (d *Discoverer) list(ctx context.Context, gvr schema.GroupVersionResource) ([]unstructured.Unstructured, error) {
	if len(d.Namespaces) == 0 {
		l, err := d.Client.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return l.Items, nil
	}
	var all []unstructured.Unstructured
	for _, ns := range d.Namespaces {
		l, err := d.Client.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		all = append(all, l.Items...)
	}
	return all, nil
}

func toWorkload(u unstructured.Unstructured, kind string, withReplicas bool) WorkloadInfo {
	w := WorkloadInfo{
		Name:         u.GetName(),
		Namespace:    u.GetNamespace(),
		Kind:         kind,
		Labels:       u.GetLabels(),
		CreationTime: u.GetCreationTimestamp().Time,
	}
	if withReplicas {
		if r, ok, _ := unstructured.NestedInt64(u.Object, "spec", "replicas"); ok {
			w.Replicas = int32(r)
		}
		if r, ok, _ := unstructured.NestedInt64(u.Object, "status", "readyReplicas"); ok {
			w.Ready = int32(r)
		}
		// DaemonSets don't have spec.replicas; use status.desiredNumberScheduled / numberReady.
		if w.Replicas == 0 {
			if r, ok, _ := unstructured.NestedInt64(u.Object, "status", "desiredNumberScheduled"); ok {
				w.Replicas = int32(r)
			}
		}
		if w.Ready == 0 {
			if r, ok, _ := unstructured.NestedInt64(u.Object, "status", "numberReady"); ok {
				w.Ready = int32(r)
			}
		}
		w.Images = extractImages(u)
	}
	return w
}

func extractImages(u unstructured.Unstructured) []string {
	containers, _, _ := unstructured.NestedSlice(u.Object, "spec", "template", "spec", "containers")
	images := make([]string, 0, len(containers))
	for _, c := range containers {
		cm, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if img, ok := cm["image"].(string); ok && img != "" {
			images = append(images, img)
		}
	}
	return images
}

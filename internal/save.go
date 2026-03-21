/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package internal

import (
	"context"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// SaveServicesToDisk writes the services config to a YAML file
func SaveServicesToDisk(services []Service, filename string) {
	config := ServiceConfig{Services: services}
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Error marshaling YAML: %v\n", err)
		return
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	if _, err = file.Write(yamlData); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("Services config written to %s\n", filename)
}

// SaveServices persists services based on the configured backend
func SaveServices(services []Service, loadFrom, configLoc, configNm string) {
	switch loadFrom {
	case "disk":
		SaveServicesToDisk(services, configLoc+"/"+configNm)
	case "cr":
		if err := CreateOrUpdateServicePortal(services, configNm, configLoc); err != nil {
			log.Printf("ERROR SAVING CR: %v", err)
		}
	default:
		log.Printf("INVALID LOAD_CONFIG_FROM VALUE: %s", loadFrom)
	}
}

// CreateOrUpdateServicePortal creates or updates a ServicePortal CR
func CreateOrUpdateServicePortal(services []Service, resourceName, namespace string) error {
	portal := &ServicePortal{
		TypeMeta: v1.TypeMeta{
			APIVersion: groupVersion.String(),
			Kind:       "ServicePortal",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      resourceName,
			Namespace: namespace,
		},
		Spec: ServicePortalSpec{
			Services: services,
		},
	}

	dynClient, err := CreateDynamicKubeConfigClient()
	if err != nil {
		return err
	}

	unstructuredConfig, err := runtime.DefaultUnstructuredConverter.ToUnstructured(portal)
	if err != nil {
		return err
	}

	resourceClient := dynClient.Resource(groupVersion.WithResource(resource)).Namespace(namespace)

	existingResource, err := resourceClient.Get(context.TODO(), portal.Name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = resourceClient.Create(context.TODO(), &unstructured.Unstructured{
				Object: unstructuredConfig,
			}, v1.CreateOptions{})
			return err
		}
		return err
	}

	unstructuredConfig["metadata"] = existingResource.Object["metadata"]
	_, err = resourceClient.Update(context.TODO(), &unstructured.Unstructured{
		Object: unstructuredConfig,
	}, v1.UpdateOptions{})
	return err
}

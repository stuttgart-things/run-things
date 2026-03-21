/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package internal

import (
	"context"
	"log"
	"os"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"gopkg.in/yaml.v2"
)

// ServicePortal CRD types
type ServicePortal struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`
	Spec          ServicePortalSpec `json:"spec,omitempty"`
}

type ServicePortalSpec struct {
	Services []Service `json:"services" yaml:"services"`
}

// CRD group/version/resource
var (
	groupVersion = schema.GroupVersion{Group: "github.stuttgart-things.com", Version: "v1"}
	resource     = "serviceportals"
)

// ReadYAMLFileFromDisk reads a YAML file from disk
func ReadYAMLFileFromDisk(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// LoadServices loads services from disk YAML or K8s CR
func LoadServices(source, configLocation, configName string) []Service {
	switch source {
	case "disk":
		yamlData, err := ReadYAMLFileFromDisk(configLocation + "/" + configName)
		if err != nil {
			log.Printf("Error reading config file: %v", err)
			return nil
		}
		var config ServiceConfig
		if err := yaml.Unmarshal(yamlData, &config); err != nil {
			log.Printf("Error parsing YAML: %v", err)
			return nil
		}
		return config.Services

	case "cr":
		portal, err := GetServicePortal(configName, configLocation)
		if err != nil {
			log.Printf("Failed to get ServicePortal: %v", err)
			return nil
		}
		return portal.Spec.Services

	default:
		log.Printf("INVALID LOAD_CONFIG_FROM VALUE: %s", source)
		return nil
	}
}

// GetServicePortal retrieves a ServicePortal CR from K8s
func GetServicePortal(resourceName, namespace string) (*ServicePortal, error) {
	dynClient, err := CreateDynamicKubeConfigClient()
	if err != nil {
		return nil, err
	}

	resourceClient := dynClient.Resource(groupVersion.WithResource(resource)).Namespace(namespace)
	unstructuredConfig, err := resourceClient.Get(context.TODO(), resourceName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var portal ServicePortal
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredConfig.Object, &portal)
	return &portal, err
}

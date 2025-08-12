// Package k8shelper provides utilities for working with Kubernetes.
package k8shelper

import (
	"os"
	"strings"
)

func getNamespace() (string, error) {
	namespaceFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	data, err := os.ReadFile(namespaceFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetEnvironment determines the environment based on the Kubernetes namespace.
func GetEnvironment() string {
	namespace, _ := getNamespace()
	environment := "local"
	if strings.Contains(namespace, "dev") {
		environment = "development"
	} else if strings.Contains(namespace, "prod") {
		environment = "production"
	} else if strings.Contains(namespace, "test") {
		environment = "testing"
	} else if strings.Contains(namespace, "staging") {
		environment = "staging"
	}
	return environment
}

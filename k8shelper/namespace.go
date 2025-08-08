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
	environment := "development"
	if strings.Contains(namespace, "prod") {
		environment = "production"
	}
	if strings.Contains(namespace, "test") {
		environment = "testing"
	}
	if strings.Contains(namespace, "staging") {
		environment = "staging"
	}
	return environment
}

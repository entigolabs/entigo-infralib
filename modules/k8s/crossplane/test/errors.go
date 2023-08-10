package test

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ObjectError struct {
	object *unstructured.Unstructured
}

func (err ObjectError) Error() string {
	return "Object is nil"
}

type ProviderNotAvailable struct {
	provider *unstructured.Unstructured
}

func (err ProviderNotAvailable) Error() string {
	status := getProviderStatus(err.provider)
	return fmt.Sprintf(
		"Provider %s is not available, healthy: %s, installed: %s", err.provider.GetName(), status["Healthy"],
		status["Installed"],
	)
}

type NewObjectError func(object *unstructured.Unstructured) error

func DefaultObjectError(object *unstructured.Unstructured) error {
	return ObjectError{object}
}

func NewProviderNotAvailable(provider *unstructured.Unstructured) error {
	return ProviderNotAvailable{provider}
}

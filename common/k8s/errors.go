package k8s

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
	status := getStatusMap(err.provider)
	return fmt.Sprintf(
		"Provider %s is not available, healthy: %s, installed: %s", err.provider.GetName(), status["Healthy"],
		status["Installed"],
	)
}

type CrossplaneObjectNotAvailable struct {
	object *unstructured.Unstructured
}

func (err CrossplaneObjectNotAvailable) Error() string {
	status := getStatusMap(err.object)
	return fmt.Sprintf(
		"%s %s is not available, ready: %s, synced: %s", err.object.GetKind(), err.object.GetName(), status["Ready"],
		status["Synced"],
	)
}

type IngressNotAvailable struct {
	ingress *unstructured.Unstructured
}

func (err IngressNotAvailable) Error() string {
	return fmt.Sprintf("Ingress %s hostname has not been set", err.ingress.GetName())
}

type NewObjectError func(object *unstructured.Unstructured) error

func DefaultObjectError(object *unstructured.Unstructured) error {
	return ObjectError{object}
}

func NewProviderNotAvailable(provider *unstructured.Unstructured) error {
	return ProviderNotAvailable{provider}
}

func NewCrossplaneObjectNotAvailable(object *unstructured.Unstructured) error {
	return CrossplaneObjectNotAvailable{object}
}

func NewIngressNotAvailable(ingress *unstructured.Unstructured) error {
	return IngressNotAvailable{ingress}
}

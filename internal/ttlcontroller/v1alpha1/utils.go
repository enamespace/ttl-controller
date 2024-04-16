package v1alpha1

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func GetObjectCreateTime(obj *unstructured.Unstructured) (time.Time, error) {

	creationTimestamp, found, err := unstructured.NestedString(obj.Object, "metadata", "creationTimestamp")
	if !found || err != nil {
		return time.Time{}, fmt.Errorf("failed to get creationTimestamp: %v", err)
	}

	parsedCreationTime, err := time.Parse(time.RFC3339, creationTimestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse creationTime: %v", err)
	}

	return parsedCreationTime.Local(), nil
}

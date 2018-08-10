package v3

import (
	"github.com/rancher/norman/condition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KontainerDriver struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the the cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
	Spec KontainerDriverSpec `json:"spec"`
	// Most recent observed status of the cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
	Status KontainerDriverStatus `json:"status"`
}

type KontainerDriverStatus struct {
	Conditions []Condition `json:"conditions"`
}

var (
	KontainerDriverConditionDownloaded condition.Cond = "Downloaded"
	KontainerDriverConditionInstalled  condition.Cond = "Installed"
	KontainerDriverConditionActive     condition.Cond = "Active"
	KontainerDriverConditionInactive   condition.Cond = "Inactive"
)

type KontainerDriverSpec struct {
	DisplayName string `json:"displayName"`
	DesiredURL  string `json:"desirdUrl" norman:"required"`
	ActualURL   string `json:"actualUrl"`
	Checksum    string `json:"checksum"`
	DesiredPort int    `json:"desiredPort" norman:"required"`
	ActualPort  int    `json:"actualPort"`
	BuiltIn     bool   `json:"builtIn"`
}

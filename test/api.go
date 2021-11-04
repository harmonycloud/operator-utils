package test

import (
	"github.com/harmonycloud/operator-utils/pkg/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MongoDBCommunity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBCommunitySpec   `json:"spec,omitempty"`
}

type MongoDBCommunitySpec struct {
	name string `json:"name"`
	StatefulSetConfiguration api.StatefulSetConfiguration `json:"statefulSet,omitempty"`
}
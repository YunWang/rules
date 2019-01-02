package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
//+genclient
//+genclient:noStatus
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Rule struct{
	metav1.TypeMeta `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec RuleSpec `json:"spec"`
}

type RuleSpec struct{
	Namespace string `json:"namespace"`
	OwnerType string `json:"ownertype"`
	OwnerName string `json:"ownername"`
	Replicas *int32 `json:"replicas"`
	nodes []NodeIp `json:"nodes"`
}


type NodeIp struct{
	Ip string `json:"ip"`
	Score string `json:"score"`
}
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RuleList struct{
	metav1.TypeMeta `json:",incline"`
	metav1.ListMeta `json:"metadata"`

	Items []Rule `json:"items"`
}
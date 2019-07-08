package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

type IP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IPSpec `json:"spec"`
}

type IPSpec struct {
	PodName     string `json:"pod_name"`
	Namespace   string `json:"namespace"`
	NodeName    string `json:"node_name"`
	MacAddress  string `json:"mac_address"`
	ContainerID string `json:"container_id"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type IPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []IP `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

type Subnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SubnetSpec `json:"spec"`
}

type SubnetSpec struct {
	Default    bool     `json:"default"`
	Protocol   string   `json:"protocol"`
	Namespaces []string `json:"namespaces"`
	CIDRBlock  string   `json:"cidr_block"`
	Gateway    string   `json:"gateway"`
	ExcludeIps []string `json:"exclude_ips"`

	GatewayType  string   `json:"gateway_type"`
	GatewayNodes []string `json:"gateway_nodes"`
	NatOutgoing  bool     `json:"nat_outgoing"`

	Private      bool     `json:"private"`
	AllowSubnets []string `json:"allow_subnets"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Subnet `json:"items"`
}

package v1


import (
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	rulecontroller "k8s.io/rules/pkg/apis/crd"
)
var SchemeGroupVersion = schema.GroupVersion{Group: rulecontroller.GroupName, Version: "v1"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error{
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&Rule{},
		&RuleList{},)
	metav1.AddToGroupVersion(scheme,SchemeGroupVersion)
	return nil
}
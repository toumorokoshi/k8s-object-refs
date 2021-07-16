package refs

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: can we get rid of this in exchange for schema.GroupVersionKind?
type GVK struct {
	Group   string
	Version string
	Kind    string
}

func GVKFromClientObject(obj client.Object) GVK {
	rawGVK := obj.GetObjectKind().GroupVersionKind()
	return GVK{
		Group:   rawGVK.Group,
		Version: rawGVK.Version,
		Kind:    rawGVK.Kind,
	}
}

func (g *GVK) ToClientObject() client.Object {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       g.Kind,
			"apiVersion": fmt.Sprintf("%v/%v", g.Group, g.Version),
		},
	}
}

type RefSubscription struct {
	Gvk            GVK
	NamespacedName types.NamespacedName
}

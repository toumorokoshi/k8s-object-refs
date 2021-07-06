package refs

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GVK struct {
	group   string
	version string
	kind    string
}

func GVKFromClientObject(obj client.Object) GVK {
	rawGVK := obj.GetObjectKind().GroupVersionKind()
	return GVK{
		group:   rawGVK.Group,
		version: rawGVK.Version,
		kind:    rawGVK.Kind,
	}
}

func (g *GVK) ToClientObject() client.Object {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       g.kind,
			"apiVersion": fmt.Sprintf("%v/%v", g.group, g.version),
		},
	}
}
